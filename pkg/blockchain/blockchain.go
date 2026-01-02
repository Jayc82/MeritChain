package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Jayc82/MeritChain/pkg/fee"
	"github.com/Jayc82/MeritChain/pkg/job"
	"github.com/Jayc82/MeritChain/pkg/reputation"
	"github.com/Jayc82/MeritChain/pkg/tokenomics"
)

const (
	// WorkerBonusPercent is the percentage of job payment awarded as bonus from worker pool
	WorkerBonusPercent = 10 // 10% bonus
	
	// ReviewerRewardDivisor is used to calculate reviewer rewards (payment / divisor / num_reviewers)
	ReviewerRewardDivisor = 20
)

// Transaction represents a transaction on the blockchain
type Transaction struct {
	ID          string
	Type        string // "transfer", "job_create", "job_accept", "job_complete", "reputation_update"
	From        string
	To          string
	Amount      int64
	Fee         int64
	Data        map[string]interface{} // Additional transaction data
	Timestamp   int64
	Signature   string
}

// Block represents a block in the blockchain
type Block struct {
	Index        int64
	Timestamp    int64
	Transactions []Transaction
	PreviousHash string
	Hash         string
	Nonce        int64
}

// Blockchain represents the main blockchain
type Blockchain struct {
	mu                sync.RWMutex
	chain             []Block
	pendingTxs        []Transaction
	reputationManager *reputation.ReputationManager
	jobManager        *job.JobManager
	feeCalculator     *fee.FeeCalculator
	tokenomics        *tokenomics.TokenomicsManager
	balances          map[string]int64 // Address -> balance for tracking on-chain income
	validatorAddress  string            // Address that receives validator rewards
}

// NewBlockchain creates a new blockchain with genesis block and tokenomics
func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		chain:             []Block{},
		pendingTxs:        []Transaction{},
		reputationManager: reputation.NewReputationManager(),
		jobManager:        job.NewJobManager(),
		feeCalculator:     fee.NewFeeCalculator(),
		tokenomics:        tokenomics.NewTokenomicsManager(),
		balances:          make(map[string]int64),
		validatorAddress:  "validator", // Default validator address
	}

	// Create validator wallet
	bc.balances["validator"] = 0

	// Create genesis block
	genesisBlock := Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PreviousHash: "0",
		Nonce:        0,
	}
	genesisBlock.Hash = bc.calculateHash(genesisBlock)
	bc.chain = append(bc.chain, genesisBlock)

	return bc
}

// calculateHash calculates the hash of a block
func (bc *Blockchain) calculateHash(block Block) string {
	record := fmt.Sprintf("%d%d%s%s%d", block.Index, block.Timestamp, block.PreviousHash, bc.serializeTransactions(block.Transactions), block.Nonce)
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// serializeTransactions converts transactions to string for hashing
func (bc *Blockchain) serializeTransactions(txs []Transaction) string {
	data, _ := json.Marshal(txs)
	return string(data)
}

// AddTransaction adds a transaction to pending transactions
func (bc *Blockchain) AddTransaction(tx Transaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Validate transaction
	if tx.From == "" && tx.Type != "job_create" {
		return errors.New("transaction must have a sender")
	}

	// Calculate and validate fee
	senderBalance := bc.balances[tx.From]
	calculatedFee, err := bc.feeCalculator.CalculateFee(senderBalance, tx.Amount)
	if err != nil {
		return fmt.Errorf("fee calculation error: %w", err)
	}

	// Apply reputation discount if available
	if rep, err := bc.reputationManager.GetReputation(tx.From); err == nil {
		calculatedFee, _ = bc.feeCalculator.CalculateFeeWithReputation(senderBalance, tx.Amount, rep.Points)
	}

	tx.Fee = calculatedFee
	tx.Timestamp = time.Now().Unix()

	bc.pendingTxs = append(bc.pendingTxs, tx)
	return nil
}

// MineBlock creates a new block with pending transactions and distributes mining rewards
func (bc *Blockchain) MineBlock() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(bc.pendingTxs) == 0 {
		return errors.New("no pending transactions to mine")
	}

	lastBlock := bc.chain[len(bc.chain)-1]
	newBlock := Block{
		Index:        lastBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: bc.pendingTxs,
		PreviousHash: lastBlock.Hash,
		Nonce:        0,
	}

	newBlock.Hash = bc.calculateHash(newBlock)

	// Process transactions
	for _, tx := range newBlock.Transactions {
		bc.processTransaction(tx)
	}

	// Mint block rewards according to tokenomics
	validatorReward, _, _, err := bc.tokenomics.MintBlockReward()
	if err != nil {
		// If we hit the supply cap, that's ok, just log it
		// Continue mining without rewards
	} else {
		// Award validator reward
		bc.balances[bc.validatorAddress] += validatorReward
		// Worker and reviewer rewards are held in their respective pools
		// They will be distributed when jobs are completed and reviews are done
	}

	bc.chain = append(bc.chain, newBlock)
	bc.pendingTxs = []Transaction{}

	return nil
}

// processTransaction processes a transaction and updates state
func (bc *Blockchain) processTransaction(tx Transaction) error {
	switch tx.Type {
	case "transfer":
		// Validate balance before transfer
		if bc.balances[tx.From] < (tx.Amount + tx.Fee) {
			return errors.New("insufficient balance for transfer")
		}
		// Update balances
		bc.balances[tx.From] -= (tx.Amount + tx.Fee)
		bc.balances[tx.To] += tx.Amount

	case "job_create":
		// Validate transaction data
		jobID, ok := tx.Data["job_id"].(string)
		if !ok {
			return errors.New("invalid job_id")
		}
		title, ok := tx.Data["title"].(string)
		if !ok {
			return errors.New("invalid title")
		}
		description, ok := tx.Data["description"].(string)
		if !ok {
			return errors.New("invalid description")
		}
		paymentFloat, ok := tx.Data["payment"].(float64)
		if !ok {
			return errors.New("invalid payment")
		}
		payment := int64(paymentFloat)
		
		reputationRewardFloat, ok := tx.Data["reputation_reward"].(float64)
		if !ok {
			return errors.New("invalid reputation_reward")
		}
		reputationReward := int64(reputationRewardFloat)

		// Validate balance before locking in escrow
		if bc.balances[tx.From] < payment {
			return errors.New("insufficient balance for job creation")
		}

		bc.jobManager.CreateJob(jobID, tx.From, title, description, payment, reputationReward, tx.Timestamp)
		bc.balances[tx.From] -= payment // Lock in escrow

	case "job_complete":
		// Validate transaction data
		jobID, ok := tx.Data["job_id"].(string)
		if !ok {
			return errors.New("invalid job_id")
		}
		
		job, err := bc.jobManager.GetJob(jobID)
		if err == nil {
			bc.jobManager.CompleteJob(jobID, tx.Timestamp)
			bc.balances[job.Worker] += job.Payment // Release escrow to worker

			// Award reputation
			bc.reputationManager.AddReputation(
				job.Worker,
				job.ReputationReward,
				"job_completed",
				fmt.Sprintf("Completed job: %s", job.Title),
				jobID,
				tx.Timestamp,
			)

			// Award worker from the worker reward pool
			// Small bonus from the pool for completing work
			workerBonus := job.Payment / WorkerBonusPercent
			if bc.tokenomics.ClaimWorkerReward(workerBonus) == nil {
				bc.balances[job.Worker] += workerBonus
			}

			// Award reviewers from the reviewer pool
			if len(job.Reviews) > 0 {
				reviewerReward := job.Payment / (ReviewerRewardDivisor * int64(len(job.Reviews)))
				for _, review := range job.Reviews {
					if bc.tokenomics.ClaimReviewerReward(reviewerReward) == nil {
						bc.balances[review.Reviewer] += reviewerReward
					}
				}
			}
		}

	case "reputation_update":
		// Validate transaction data
		address, ok := tx.Data["address"].(string)
		if !ok {
			return errors.New("invalid address")
		}
		pointsFloat, ok := tx.Data["points"].(float64)
		if !ok {
			return errors.New("invalid points")
		}
		points := int64(pointsFloat)
		
		eventType, ok := tx.Data["event_type"].(string)
		if !ok {
			return errors.New("invalid event_type")
		}
		description, ok := tx.Data["description"].(string)
		if !ok {
			return errors.New("invalid description")
		}

		if points > 0 {
			bc.reputationManager.AddReputation(address, points, eventType, description, tx.ID, tx.Timestamp)
		} else {
			bc.reputationManager.RemoveReputation(address, -points, eventType, description, tx.ID, tx.Timestamp)
		}
	}

	return nil
}

// CreateWallet creates a new reputation wallet for an address
func (bc *Blockchain) CreateWallet(address string) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if _, exists := bc.balances[address]; exists {
		return errors.New("wallet already exists")
	}

	bc.balances[address] = 0
	return bc.reputationManager.CreateWallet(address)
}

// GetBalance returns the balance for an address
func (bc *Blockchain) GetBalance(address string) int64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.balances[address]
}

// GetReputation returns reputation for an address
func (bc *Blockchain) GetReputation(address string) (*reputation.Reputation, error) {
	return bc.reputationManager.GetReputation(address)
}

// GetJob returns a job by ID
func (bc *Blockchain) GetJob(jobID string) (*job.Job, error) {
	return bc.jobManager.GetJob(jobID)
}

// GetChain returns the entire blockchain
func (bc *Blockchain) GetChain() []Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.chain
}

// GetLatestBlock returns the latest block
func (bc *Blockchain) GetLatestBlock() Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.chain[len(bc.chain)-1]
}

// IsValid validates the entire blockchain
func (bc *Blockchain) IsValid() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for i := 1; i < len(bc.chain); i++ {
		currentBlock := bc.chain[i]
		previousBlock := bc.chain[i-1]

		if currentBlock.Hash != bc.calculateHash(currentBlock) {
			return false
		}

		if currentBlock.PreviousHash != previousBlock.Hash {
			return false
		}
	}

	return true
}

// GetReputationManager returns the reputation manager
func (bc *Blockchain) GetReputationManager() *reputation.ReputationManager {
	return bc.reputationManager
}

// GetJobManager returns the job manager
func (bc *Blockchain) GetJobManager() *job.JobManager {
	return bc.jobManager
}

// GetFeeCalculator returns the fee calculator
func (bc *Blockchain) GetFeeCalculator() *fee.FeeCalculator {
	return bc.feeCalculator
}

// GetTokenomics returns the tokenomics manager
func (bc *Blockchain) GetTokenomics() *tokenomics.TokenomicsManager {
	return bc.tokenomics
}

// DistributeCommunityFunds distributes coins from the community bootstrap pool
func (bc *Blockchain) DistributeCommunityFunds(address string, amount int64) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	err := bc.tokenomics.AllocateCommunityFunds(amount)
	if err != nil {
		return err
	}

	bc.balances[address] += amount
	return nil
}

// DistributeProtocolFunds distributes coins from the protocol reserve
func (bc *Blockchain) DistributeProtocolFunds(address string, amount int64) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	err := bc.tokenomics.AllocateProtocolFunds(amount)
	if err != nil {
		return err
	}

	bc.balances[address] += amount
	return nil
}

// GetTotalSupply returns the total supply cap
func (bc *Blockchain) GetTotalSupply() int64 {
	return tokenomics.TotalSupplyCap
}

// GetCirculatingSupply returns the circulating supply
func (bc *Blockchain) GetCirculatingSupply() int64 {
	return bc.tokenomics.GetCirculatingSupply()
}

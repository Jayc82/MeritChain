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
	balances          map[string]int64 // Address -> balance for tracking on-chain income
}

// NewBlockchain creates a new blockchain with genesis block
func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		chain:             []Block{},
		pendingTxs:        []Transaction{},
		reputationManager: reputation.NewReputationManager(),
		jobManager:        job.NewJobManager(),
		feeCalculator:     fee.NewFeeCalculator(),
		balances:          make(map[string]int64),
	}

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

// MineBlock creates a new block with pending transactions
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

	bc.chain = append(bc.chain, newBlock)
	bc.pendingTxs = []Transaction{}

	return nil
}

// processTransaction processes a transaction and updates state
func (bc *Blockchain) processTransaction(tx Transaction) error {
	switch tx.Type {
	case "transfer":
		// Update balances
		bc.balances[tx.From] -= (tx.Amount + tx.Fee)
		bc.balances[tx.To] += tx.Amount

	case "job_create":
		// Create job
		jobID := tx.Data["job_id"].(string)
		title := tx.Data["title"].(string)
		description := tx.Data["description"].(string)
		payment := int64(tx.Data["payment"].(float64))
		reputationReward := int64(tx.Data["reputation_reward"].(float64))

		bc.jobManager.CreateJob(jobID, tx.From, title, description, payment, reputationReward, tx.Timestamp)
		bc.balances[tx.From] -= payment // Lock in escrow

	case "job_complete":
		// Complete job and award reputation
		jobID := tx.Data["job_id"].(string)
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
		}

	case "reputation_update":
		// Manual reputation update (e.g., from peer review)
		address := tx.Data["address"].(string)
		points := int64(tx.Data["points"].(float64))
		eventType := tx.Data["event_type"].(string)
		description := tx.Data["description"].(string)

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

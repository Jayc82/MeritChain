package tokenomics

import (
	"errors"
	"sync"
)

// Constants for the tokenomics model
const (
	// Total supply cap: 21 million coins
	TotalSupplyCap int64 = 21_000_000

	// Genesis allocation percentages
	GenesisReservePercent      = 70 // Reserved for future work + validator rewards
	GenesisCommunityPercent    = 20 // Community bootstrap (early workers, testers)
	GenesisProtocolPercent     = 10 // Protocol reserve (development, audits)

	// Block reward split percentages
	ValidatorRewardPercent   = 60 // Network security
	WorkerRewardPercent      = 30 // Job completion pool
	ReviewerRewardPercent    = 10 // Peer reviewers & dispute jurors

	// Halving parameters
	BlocksPerHalving     = 210_000 // Approximately every 4 years (similar to Bitcoin)
	InitialBlockReward   = 50      // Starting block reward in coins
)

// TokenomicsManager manages coin supply, distribution, and rewards
type TokenomicsManager struct {
	mu                    sync.RWMutex
	totalEmitted          int64 // Total coins emitted through block rewards
	totalSupplyCap        int64
	genesisReserve        int64 // Locked for future emission
	communityPool         int64 // Available for distribution
	protocolReserve       int64 // Development and emergency funds
	workerRewardPool      int64 // Accumulated rewards for workers
	reviewerRewardPool    int64 // Accumulated rewards for reviewers
	blockHeight           int64 // Current block height
	emissionSchedule      map[int64]int64 // Block height -> reward amount
}

// NewTokenomicsManager creates a new tokenomics manager with genesis allocation
func NewTokenomicsManager() *TokenomicsManager {
	tm := &TokenomicsManager{
		totalEmitted:     0,
		totalSupplyCap:   TotalSupplyCap,
		blockHeight:      0,
		emissionSchedule: make(map[int64]int64),
	}

	// Calculate genesis allocations
	tm.genesisReserve = (TotalSupplyCap * GenesisReservePercent) / 100
	tm.communityPool = (TotalSupplyCap * GenesisCommunityPercent) / 100
	tm.protocolReserve = (TotalSupplyCap * GenesisProtocolPercent) / 100

	// These are allocated but not yet emitted (locked/reserved)
	// totalEmitted tracks only what's been distributed through mining

	return tm
}

// GetCurrentBlockReward calculates the block reward for the current block height
// Implements halving mechanism: reward decreases every BlocksPerHalving blocks
func (tm *TokenomicsManager) GetCurrentBlockReward() int64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Calculate number of halvings that have occurred
	halvings := tm.blockHeight / BlocksPerHalving
	
	// If we've had more than 64 halvings, reward is effectively 0
	if halvings >= 64 {
		return 0
	}

	// Calculate reward: initial_reward / (2 ^ halvings)
	reward := int64(InitialBlockReward >> halvings)
	
	// Check if minting this reward would exceed the cap
	if tm.totalEmitted + reward > tm.totalSupplyCap {
		// Only mint what's left under the cap
		remaining := tm.totalSupplyCap - tm.totalEmitted
		if remaining < 0 {
			return 0
		}
		return remaining
	}

	return reward
}

// MintBlockReward mints new coins for a block and distributes them according to the model
// Returns (validatorReward, workerPoolReward, reviewerPoolReward, error)
func (tm *TokenomicsManager) MintBlockReward() (int64, int64, int64, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Get the reward for current block
	blockReward := tm.getCurrentBlockRewardUnlocked()
	
	if blockReward == 0 {
		return 0, 0, 0, nil // No more rewards to mint
	}

	// Check supply cap
	if tm.totalEmitted + blockReward > tm.totalSupplyCap {
		return 0, 0, 0, errors.New("cannot exceed total supply cap")
	}

	// Split the reward according to percentages
	validatorReward := (blockReward * ValidatorRewardPercent) / 100
	workerReward := (blockReward * WorkerRewardPercent) / 100
	reviewerReward := (blockReward * ReviewerRewardPercent) / 100

	// Ensure total doesn't exceed block reward due to rounding
	totalDistributed := validatorReward + workerReward + reviewerReward
	if totalDistributed > blockReward {
		// Adjust validator reward down by the difference
		validatorReward -= (totalDistributed - blockReward)
	}

	// Add to pools
	tm.workerRewardPool += workerReward
	tm.reviewerRewardPool += reviewerReward

	// Update total minted
	tm.totalEmitted += blockReward
	tm.blockHeight++

	return validatorReward, workerReward, reviewerReward, nil
}

// getCurrentBlockRewardUnlocked is the internal version without locking
func (tm *TokenomicsManager) getCurrentBlockRewardUnlocked() int64 {
	halvings := tm.blockHeight / BlocksPerHalving
	if halvings >= 64 {
		return 0
	}
	reward := int64(InitialBlockReward >> halvings)
	if tm.totalEmitted + reward > tm.totalSupplyCap {
		remaining := tm.totalSupplyCap - tm.totalEmitted
		if remaining < 0 {
			return 0
		}
		return remaining
	}
	return reward
}

// AllocateCommunityFunds distributes coins from the community bootstrap pool
func (tm *TokenomicsManager) AllocateCommunityFunds(amount int64) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if tm.communityPool < amount {
		return errors.New("insufficient funds in community pool")
	}

	tm.communityPool -= amount
	return nil
}

// AllocateProtocolFunds distributes coins from the protocol reserve
func (tm *TokenomicsManager) AllocateProtocolFunds(amount int64) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if tm.protocolReserve < amount {
		return errors.New("insufficient funds in protocol reserve")
	}

	tm.protocolReserve -= amount
	return nil
}

// ClaimWorkerReward allows claiming from the worker reward pool
func (tm *TokenomicsManager) ClaimWorkerReward(amount int64) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if tm.workerRewardPool < amount {
		return errors.New("insufficient funds in worker reward pool")
	}

	tm.workerRewardPool -= amount
	return nil
}

// ClaimReviewerReward allows claiming from the reviewer reward pool
func (tm *TokenomicsManager) ClaimReviewerReward(amount int64) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if tm.reviewerRewardPool < amount {
		return errors.New("insufficient funds in reviewer reward pool")
	}

	tm.reviewerRewardPool -= amount
	return nil
}

// GetStats returns current tokenomics statistics
func (tm *TokenomicsManager) GetStats() TokenomicsStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Circulating supply = emitted through mining + distributed from pools
	// (Genesis reserve is locked and only released through mining rewards)
	initialCommunity := (TotalSupplyCap * GenesisCommunityPercent) / 100
	initialProtocol := (TotalSupplyCap * GenesisProtocolPercent) / 100
	distributedFromPools := (initialCommunity - tm.communityPool) + (initialProtocol - tm.protocolReserve)
	
	circulatingSupply := tm.totalEmitted + distributedFromPools
	percentEmitted := float64(tm.totalEmitted) / float64(tm.genesisReserve) * 100 // % of reserve emitted

	return TokenomicsStats{
		TotalSupplyCap:     tm.totalSupplyCap,
		TotalMinted:        tm.totalEmitted,
		CirculatingSupply:  circulatingSupply,
		PercentMinted:      percentEmitted,
		GenesisReserve:     tm.genesisReserve,
		CommunityPool:      tm.communityPool,
		ProtocolReserve:    tm.protocolReserve,
		WorkerRewardPool:   tm.workerRewardPool,
		ReviewerRewardPool: tm.reviewerRewardPool,
		CurrentBlockHeight: tm.blockHeight,
		CurrentBlockReward: tm.getCurrentBlockRewardUnlocked(),
		NextHalvingBlock:   ((tm.blockHeight / BlocksPerHalving) + 1) * BlocksPerHalving,
	}
}

// TokenomicsStats represents the current state of the tokenomics
type TokenomicsStats struct {
	TotalSupplyCap     int64
	TotalMinted        int64
	CirculatingSupply  int64
	PercentMinted      float64
	GenesisReserve     int64
	CommunityPool      int64
	ProtocolReserve    int64
	WorkerRewardPool   int64
	ReviewerRewardPool int64
	CurrentBlockHeight int64
	CurrentBlockReward int64
	NextHalvingBlock   int64
}

// GetEmissionSchedule calculates the emission schedule for the next N blocks
func (tm *TokenomicsManager) GetEmissionSchedule(blocks int64) []EmissionInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	schedule := make([]EmissionInfo, 0)
	tempHeight := tm.blockHeight
	tempMinted := tm.totalEmitted

	for i := int64(0); i < blocks; i++ {
		halvings := tempHeight / BlocksPerHalving
		if halvings >= 64 {
			break
		}

		reward := int64(InitialBlockReward >> halvings)
		if tempMinted + reward > tm.totalSupplyCap {
			remaining := tm.totalSupplyCap - tempMinted
			if remaining <= 0 {
				break
			}
			reward = remaining
		}

		schedule = append(schedule, EmissionInfo{
			BlockHeight: tempHeight,
			Reward:      reward,
			TotalMinted: tempMinted + reward,
		})

		tempHeight++
		tempMinted += reward
	}

	return schedule
}

// EmissionInfo represents emission information for a specific block
type EmissionInfo struct {
	BlockHeight int64
	Reward      int64
	TotalMinted int64
}

// CalculateYearsToFullEmission estimates years until full supply is minted
// Assumes average block time in seconds
func CalculateYearsToFullEmission(avgBlockTimeSeconds float64) float64 {
	totalBlocks := int64(0)
	totalEmitted := int64(0)
	blockHeight := int64(0)

	for totalEmitted < TotalSupplyCap {
		halvings := blockHeight / BlocksPerHalving
		if halvings >= 64 {
			break
		}

		reward := int64(InitialBlockReward >> halvings)
		if totalEmitted + reward > TotalSupplyCap {
			break
		}

		totalEmitted += reward
		blockHeight++
		totalBlocks++

		// Safety limit to prevent infinite loop
		if totalBlocks > 10_000_000 {
			break
		}
	}

	totalSeconds := float64(totalBlocks) * avgBlockTimeSeconds
	years := totalSeconds / (365.25 * 24 * 60 * 60)

	return years
}

// GetBlockHeight returns current block height
func (tm *TokenomicsManager) GetBlockHeight() int64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.blockHeight
}

// GetCirculatingSupply returns the circulating supply (emitted + distributed from pools)
func (tm *TokenomicsManager) GetCirculatingSupply() int64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	initialCommunity := (TotalSupplyCap * GenesisCommunityPercent) / 100
	initialProtocol := (TotalSupplyCap * GenesisProtocolPercent) / 100
	distributedFromPools := (initialCommunity - tm.communityPool) + (initialProtocol - tm.protocolReserve)
	
	return tm.totalEmitted + distributedFromPools
}

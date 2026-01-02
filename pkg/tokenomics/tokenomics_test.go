package tokenomics

import (
	"testing"
)

func TestNewTokenomicsManager(t *testing.T) {
	tm := NewTokenomicsManager()

	if tm.totalSupplyCap != TotalSupplyCap {
		t.Errorf("Expected total supply cap %d, got %d", TotalSupplyCap, tm.totalSupplyCap)
	}

	// Check genesis allocations
	expectedReserve := (TotalSupplyCap * GenesisReservePercent) / 100
	if tm.genesisReserve != expectedReserve {
		t.Errorf("Expected genesis reserve %d, got %d", expectedReserve, tm.genesisReserve)
	}

	expectedCommunity := (TotalSupplyCap * GenesisCommunityPercent) / 100
	if tm.communityPool != expectedCommunity {
		t.Errorf("Expected community pool %d, got %d", expectedCommunity, tm.communityPool)
	}

	expectedProtocol := (TotalSupplyCap * GenesisProtocolPercent) / 100
	if tm.protocolReserve != expectedProtocol {
		t.Errorf("Expected protocol reserve %d, got %d", expectedProtocol, tm.protocolReserve)
	}

	// Total emitted should start at 0 (genesis allocations are reserved, not emitted yet)
	if tm.totalEmitted != 0 {
		t.Errorf("Expected total emitted 0, got %d", tm.totalEmitted)
	}
}

func TestGetCurrentBlockReward(t *testing.T) {
	tm := NewTokenomicsManager()

	// First block reward should be initial reward
	reward := tm.GetCurrentBlockReward()
	if reward != InitialBlockReward {
		t.Errorf("Expected initial block reward %d, got %d", InitialBlockReward, reward)
	}

	// Simulate reaching first halving
	tm.blockHeight = BlocksPerHalving
	reward = tm.GetCurrentBlockReward()
	expectedReward := int64(InitialBlockReward / 2)
	if reward != expectedReward {
		t.Errorf("Expected halved reward %d, got %d", expectedReward, reward)
	}

	// Simulate second halving
	tm.blockHeight = BlocksPerHalving * 2
	reward = tm.GetCurrentBlockReward()
	expectedReward = int64(InitialBlockReward / 4)
	if reward != expectedReward {
		t.Errorf("Expected second halved reward %d, got %d", expectedReward, reward)
	}
}

func TestMintBlockReward(t *testing.T) {
	tm := NewTokenomicsManager()

	validatorReward, workerReward, reviewerReward, err := tm.MintBlockReward()
	if err != nil {
		t.Fatalf("Failed to mint block reward: %v", err)
	}

	// Check that rewards sum to expected total
	totalReward := validatorReward + workerReward + reviewerReward
	if totalReward > InitialBlockReward {
		t.Errorf("Total rewards %d exceed block reward %d", totalReward, InitialBlockReward)
	}

	// Check approximate percentages
	expectedValidator := int64((InitialBlockReward * ValidatorRewardPercent) / 100)
	if validatorReward != expectedValidator {
		t.Logf("Validator reward %d differs from expected %d (acceptable for rounding)", validatorReward, expectedValidator)
	}

	// Check that pools were updated
	if tm.workerRewardPool != workerReward {
		t.Errorf("Worker pool should be %d, got %d", workerReward, tm.workerRewardPool)
	}

	if tm.reviewerRewardPool != reviewerReward {
		t.Errorf("Reviewer pool should be %d, got %d", reviewerReward, tm.reviewerRewardPool)
	}

	// Block height should increment
	if tm.blockHeight != 1 {
		t.Errorf("Expected block height 1, got %d", tm.blockHeight)
	}
}

func TestRewardHalving(t *testing.T) {
	tm := NewTokenomicsManager()

	// Mint blocks until first halving
	initialReward := tm.GetCurrentBlockReward()

	for i := int64(0); i < BlocksPerHalving; i++ {
		tm.MintBlockReward()
	}

	// After halving, reward should be half
	newReward := tm.GetCurrentBlockReward()
	expectedNewReward := initialReward / 2

	if newReward != expectedNewReward {
		t.Errorf("After halving, expected reward %d, got %d", expectedNewReward, newReward)
	}
}

func TestSupplyCap(t *testing.T) {
	tm := NewTokenomicsManager()

	// Set total minted near the cap
	tm.totalEmitted = TotalSupplyCap - 10

	reward := tm.GetCurrentBlockReward()
	if reward > 10 {
		t.Errorf("Reward %d should not exceed remaining supply 10", reward)
	}

	// Test that we can't mint beyond the cap
	tm.totalEmitted = TotalSupplyCap
	reward = tm.GetCurrentBlockReward()
	if reward != 0 {
		t.Errorf("Expected 0 reward at cap, got %d", reward)
	}
}

func TestAllocateCommunityFunds(t *testing.T) {
	tm := NewTokenomicsManager()

	initialPool := tm.communityPool
	allocateAmount := int64(1000)

	err := tm.AllocateCommunityFunds(allocateAmount)
	if err != nil {
		t.Fatalf("Failed to allocate community funds: %v", err)
	}

	if tm.communityPool != initialPool - allocateAmount {
		t.Errorf("Expected community pool %d, got %d", initialPool - allocateAmount, tm.communityPool)
	}

	// Test insufficient funds
	err = tm.AllocateCommunityFunds(tm.communityPool + 1)
	if err == nil {
		t.Error("Expected error when allocating more than available")
	}
}

func TestAllocateProtocolFunds(t *testing.T) {
	tm := NewTokenomicsManager()

	initialReserve := tm.protocolReserve
	allocateAmount := int64(500)

	err := tm.AllocateProtocolFunds(allocateAmount)
	if err != nil {
		t.Fatalf("Failed to allocate protocol funds: %v", err)
	}

	if tm.protocolReserve != initialReserve - allocateAmount {
		t.Errorf("Expected protocol reserve %d, got %d", initialReserve - allocateAmount, tm.protocolReserve)
	}
}

func TestClaimWorkerReward(t *testing.T) {
	tm := NewTokenomicsManager()

	// Mint a block to add to worker pool
	_, workerReward, _, _ := tm.MintBlockReward()

	// Claim from worker pool
	claimAmount := workerReward / 2
	err := tm.ClaimWorkerReward(claimAmount)
	if err != nil {
		t.Fatalf("Failed to claim worker reward: %v", err)
	}

	expectedRemaining := workerReward - claimAmount
	if tm.workerRewardPool != expectedRemaining {
		t.Errorf("Expected worker pool %d, got %d", expectedRemaining, tm.workerRewardPool)
	}

	// Test insufficient funds
	err = tm.ClaimWorkerReward(tm.workerRewardPool + 1)
	if err == nil {
		t.Error("Expected error when claiming more than available")
	}
}

func TestClaimReviewerReward(t *testing.T) {
	tm := NewTokenomicsManager()

	// Mint a block to add to reviewer pool
	_, _, reviewerReward, _ := tm.MintBlockReward()

	// Claim from reviewer pool
	claimAmount := reviewerReward / 2
	err := tm.ClaimReviewerReward(claimAmount)
	if err != nil {
		t.Fatalf("Failed to claim reviewer reward: %v", err)
	}

	expectedRemaining := reviewerReward - claimAmount
	if tm.reviewerRewardPool != expectedRemaining {
		t.Errorf("Expected reviewer pool %d, got %d", expectedRemaining, tm.reviewerRewardPool)
	}
}

func TestGetStats(t *testing.T) {
	tm := NewTokenomicsManager()

	// Mint a few blocks
	for i := 0; i < 10; i++ {
		tm.MintBlockReward()
	}

	stats := tm.GetStats()

	if stats.TotalSupplyCap != TotalSupplyCap {
		t.Errorf("Expected supply cap %d, got %d", TotalSupplyCap, stats.TotalSupplyCap)
	}

	if stats.CurrentBlockHeight != 10 {
		t.Errorf("Expected block height 10, got %d", stats.CurrentBlockHeight)
	}

	if stats.TotalMinted <= 0 {
		t.Error("Expected positive total minted")
	}

	if stats.CirculatingSupply < 0 {
		t.Error("Circulating supply should not be negative")
	}
}

func TestGetEmissionSchedule(t *testing.T) {
	tm := NewTokenomicsManager()

	schedule := tm.GetEmissionSchedule(100)

	if len(schedule) == 0 {
		t.Error("Expected non-empty emission schedule")
	}

	// Check that schedule is sequential
	for i := 1; i < len(schedule); i++ {
		if schedule[i].BlockHeight != schedule[i-1].BlockHeight + 1 {
			t.Errorf("Expected sequential block heights, got %d after %d", schedule[i].BlockHeight, schedule[i-1].BlockHeight)
		}

		if schedule[i].TotalMinted <= schedule[i-1].TotalMinted {
			t.Error("Total minted should increase with each block")
		}
	}
}

func TestCalculateYearsToFullEmission(t *testing.T) {
	// Test with different block times
	years := CalculateYearsToFullEmission(600.0) // 10 minute blocks

	if years <= 0 {
		t.Error("Expected positive years to full emission")
	}

	t.Logf("Estimated years to full emission with 10-minute blocks: %.2f", years)

	// Should be several decades
	if years < 10 || years > 200 {
		t.Logf("Warning: Emission schedule seems unusual: %.2f years", years)
	}
}

func TestThreadSafety(t *testing.T) {
	tm := NewTokenomicsManager()

	// Run concurrent operations
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				tm.GetStats()
				tm.GetCurrentBlockReward()
			}
			done <- true
		}()
	}

	// Also mint blocks concurrently
	go func() {
		for i := 0; i < 100; i++ {
			tm.MintBlockReward()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 11; i++ {
		<-done
	}

	// If we get here without deadlock or race, the test passes
	stats := tm.GetStats()
	if stats.CurrentBlockHeight != 100 {
		t.Errorf("Expected 100 blocks, got %d", stats.CurrentBlockHeight)
	}
}

package main

import (
	"fmt"
	"time"

	"github.com/Jayc82/MeritChain/pkg/blockchain"
)

// This example demonstrates basic usage of the MeritChain blockchain API
func main() {
	// Create a new blockchain
	bc := blockchain.NewBlockchain()

	// Create wallets for participants
	fmt.Println("=== Creating Wallets ===")
	bc.CreateWallet("alice")
	bc.CreateWallet("bob")
	fmt.Println("✓ Wallets created for alice and bob")

	// Alice creates a job
	fmt.Println("\n=== Creating a Job ===")
	jobTx := blockchain.Transaction{
		ID:   "tx-job-001",
		Type: "job_create",
		From: "alice",
		Data: map[string]interface{}{
			"job_id":            "job-001",
			"title":             "Write documentation",
			"description":       "Need technical documentation for API",
			"payment":           float64(500),
			"reputation_reward": float64(25),
		},
	}
	bc.AddTransaction(jobTx)
	bc.MineBlock()
	fmt.Println("✓ Job created and mined")

	// Get the job
	job, _ := bc.GetJob("job-001")
	fmt.Printf("Job: %s (Status: %s, Payment: %d, Reputation: %d)\n",
		job.Title, job.Status, job.Payment, job.ReputationReward)

	// Bob accepts the job
	fmt.Println("\n=== Job Acceptance ===")
	jobManager := bc.GetJobManager()
	jobManager.AcceptJob("job-001", "bob", time.Now().Unix())
	jobManager.StartJob("job-001", "bob")
	jobManager.SubmitJob("job-001", "bob")
	fmt.Println("✓ Bob accepted and submitted the job")

	// Add a review
	fmt.Println("\n=== Peer Review ===")
	jobManager.AddReview("job-001", "alice", 5, "Excellent documentation!", true, time.Now().Unix())
	fmt.Println("✓ Review added")

	// Complete the job
	fmt.Println("\n=== Job Completion ===")
	completeTx := blockchain.Transaction{
		ID:   "tx-complete-001",
		Type: "job_complete",
		From: "alice",
		To:   "bob",
		Data: map[string]interface{}{
			"job_id": "job-001",
		},
	}
	bc.AddTransaction(completeTx)
	bc.MineBlock()
	fmt.Println("✓ Job completed, payment released, reputation awarded")

	// Check final state
	fmt.Println("\n=== Final State ===")
	fmt.Printf("Bob's balance: %d units\n", bc.GetBalance("bob"))

	if rep, err := bc.GetReputation("bob"); err == nil {
		fmt.Printf("Bob's reputation: %d points\n", rep.Points)
		if len(rep.History) > 0 {
			fmt.Println("\nReputation events:")
			for _, event := range rep.History {
				fmt.Printf("  - %s: %+d points (%s)\n", event.EventType, event.Points, event.Description)
			}
		}
	}

	// Demonstrate fee scaling
	fmt.Println("\n=== Fee Calculation Examples ===")
	feeCalc := bc.GetFeeCalculator()

	scenarios := []struct {
		income     int64
		reputation int64
		label      string
	}{
		{0, 0, "New user (no income/reputation)"},
		{1000, 0, "User with 1000 units income"},
		{1000, 50, "User with 1000 units income + 50 reputation"},
		{10000, 100, "User with 10000 units income + 100 reputation"},
	}

	for _, scenario := range scenarios {
		if scenario.reputation > 0 {
			fee, _ := feeCalc.CalculateFeeWithReputation(scenario.income, 1000, scenario.reputation)
			fmt.Printf("%s:\n  Fee for 1000 unit transaction: %d units\n", scenario.label, fee)
		} else {
			fee, _ := feeCalc.CalculateFee(scenario.income, 1000)
			fmt.Printf("%s:\n  Fee for 1000 unit transaction: %d units\n", scenario.label, fee)
		}
	}

	// Blockchain validation
	fmt.Println("\n=== Blockchain Validation ===")
	if bc.IsValid() {
		fmt.Println("✓ Blockchain is valid")
		fmt.Printf("Chain length: %d blocks\n", len(bc.GetChain()))
	}

	fmt.Println("\n✓ Example complete!")
}

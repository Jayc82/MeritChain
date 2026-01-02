package main

import (
	"fmt"
	"time"

	"github.com/Jayc82/MeritChain/pkg/blockchain"
)

// This example demonstrates basic usage of the MeritChain blockchain API
// Note: This example uses the managers directly for demonstration purposes.
// In production, all operations would go through blockchain transactions.
func main() {
	// Create a new blockchain
	bc := blockchain.NewBlockchain()

	// Create wallets for participants
	fmt.Println("=== Creating Wallets ===")
	bc.CreateWallet("alice")
	bc.CreateWallet("bob")
	fmt.Println("✓ Wallets created for alice and bob")

	// Note: In a real blockchain, users would receive initial funds through mining or distribution
	// For this demo, we use the managers directly to show the core functionality
	fmt.Println("\n=== Creating a Job ===")
	fmt.Println("Note: Using JobManager directly for demonstration")
	fmt.Println("In production, this would be done through blockchain transactions")
	
	jobManager := bc.GetJobManager()
	job, err := jobManager.CreateJob("job-001", "alice", "Write documentation", "Need technical documentation for API", 500, 25, time.Now().Unix())
	if err != nil {
		fmt.Printf("Error creating job: %v\n", err)
		return
	}
	fmt.Printf("✓ Job created: %s (Status: %s, Payment: %d, Reputation: %d)\n",
		job.Title, job.Status, job.Payment, job.ReputationReward)

	// Bob accepts the job
	fmt.Println("\n=== Job Acceptance ===")
	err = jobManager.AcceptJob("job-001", "bob", time.Now().Unix())
	if err != nil {
		fmt.Printf("Error accepting job: %v\n", err)
		return
	}
	err = jobManager.StartJob("job-001", "bob")
	if err != nil {
		fmt.Printf("Error starting job: %v\n", err)
		return
	}
	err = jobManager.SubmitJob("job-001", "bob")
	if err != nil {
		fmt.Printf("Error submitting job: %v\n", err)
		return
	}
	fmt.Println("✓ Bob accepted, started, and submitted the job")

	// Add a review
	fmt.Println("\n=== Peer Review ===")
	err = jobManager.AddReview("job-001", "alice", 5, "Excellent documentation!", true, time.Now().Unix())
	if err != nil {
		fmt.Printf("Error adding review: %v\n", err)
		return
	}
	fmt.Println("✓ Review added (5 stars, approved)")

	// Complete the job
	fmt.Println("\n=== Job Completion ===")
	err = jobManager.CompleteJob("job-001", time.Now().Unix())
	if err != nil {
		fmt.Printf("Error completing job: %v\n", err)
		return
	}
	fmt.Println("✓ Job completed, escrow released")

	// Award reputation
	repManager := bc.GetReputationManager()
	err = repManager.AddReputation("bob", 25, "job_completed", "Completed job: Write documentation", "job-001", time.Now().Unix())
	if err != nil {
		fmt.Printf("Error adding reputation: %v\n", err)
		return
	}
	fmt.Println("✓ Reputation awarded to bob")

	// Check final state
	fmt.Println("\n=== Final State ===")
	if rep, err := bc.GetReputation("bob"); err == nil {
		fmt.Printf("Bob's reputation: %d points\n", rep.Points)
		if len(rep.History) > 0 {
			fmt.Println("\nReputation history:")
			for _, event := range rep.History {
				fmt.Printf("  - %s: %+d points (%s)\n", event.EventType, event.Points, event.Description)
			}
		}
	}

	// Get job details
	completedJob, err := jobManager.GetJob("job-001")
	if err == nil {
		fmt.Printf("\nJob status: %s\n", completedJob.Status)
		fmt.Printf("Reviews: %d review(s)\n", len(completedJob.Reviews))
		if len(completedJob.Reviews) > 0 {
			review := completedJob.Reviews[0]
			fmt.Printf("  - Rating: %d/5 stars, Comment: \"%s\"\n", review.Rating, review.Comment)
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
		{1000, 25, "User with 1000 units income + 25 reputation (like bob)"},
		{10000, 100, "User with 10000 units income + 100 reputation"},
	}

	for _, scenario := range scenarios {
		if scenario.reputation > 0 {
			fee, _ := feeCalc.CalculateFeeWithReputation(scenario.income, 1000, scenario.reputation)
			fmt.Printf("%-50s: %d units\n", scenario.label, fee)
		} else {
			fee, _ := feeCalc.CalculateFee(scenario.income, 1000)
			fmt.Printf("%-50s: %d units\n", scenario.label, fee)
		}
	}

	// Blockchain validation
	fmt.Println("\n=== Blockchain Validation ===")
	if bc.IsValid() {
		fmt.Println("✓ Blockchain is valid")
		fmt.Printf("Chain length: %d block(s)\n", len(bc.GetChain()))
	}

	fmt.Println("\n✓ Example complete!")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("  • Reputation is earned through completed work (non-transferable)")
	fmt.Println("  • Jobs use escrow to protect both parties")
	fmt.Println("  • Peer reviews ensure quality")
	fmt.Println("  • Transaction fees scale with income (more equitable access)")
	fmt.Println("  • Reputation earns fee discounts (rewards honest participation)")
}


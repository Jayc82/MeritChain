package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Jayc82/MeritChain/pkg/blockchain"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	bc := blockchain.NewBlockchain()

	command := os.Args[1]

	switch command {
	case "wallet":
		if len(os.Args) < 3 {
			fmt.Println("Usage: meritchain wallet <address>")
			os.Exit(1)
		}
		createWallet(bc, os.Args[2])

	case "job":
		if len(os.Args) < 8 {
			fmt.Println("Usage: meritchain job <id> <poster> <title> <description> <payment> <reputation>")
			os.Exit(1)
		}
		createJob(bc, os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6], os.Args[7])

	case "transfer":
		if len(os.Args) < 5 {
			fmt.Println("Usage: meritchain transfer <from> <to> <amount>")
			os.Exit(1)
		}
		transfer(bc, os.Args[2], os.Args[3], os.Args[4])

	case "mine":
		mine(bc)

	case "balance":
		if len(os.Args) < 3 {
			fmt.Println("Usage: meritchain balance <address>")
			os.Exit(1)
		}
		getBalance(bc, os.Args[2])

	case "reputation":
		if len(os.Args) < 3 {
			fmt.Println("Usage: meritchain reputation <address>")
			os.Exit(1)
		}
		getReputation(bc, os.Args[2])

	case "chain":
		printChain(bc)

	case "demo":
		runDemo(bc)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("MeritChain - A blockchain where reputation replaces identity and work replaces speculation")
	fmt.Println("\nUsage:")
	fmt.Println("  meritchain wallet <address>                                    - Create a new wallet")
	fmt.Println("  meritchain job <id> <poster> <title> <desc> <payment> <rep>  - Create a job")
	fmt.Println("  meritchain transfer <from> <to> <amount>                      - Transfer funds")
	fmt.Println("  meritchain mine                                                - Mine pending transactions")
	fmt.Println("  meritchain balance <address>                                   - Check balance")
	fmt.Println("  meritchain reputation <address>                                - Check reputation")
	fmt.Println("  meritchain chain                                               - Print blockchain")
	fmt.Println("  meritchain demo                                                - Run demonstration")
}

func createWallet(bc *blockchain.Blockchain, address string) {
	err := bc.CreateWallet(address)
	if err != nil {
		fmt.Printf("Error creating wallet: %v\n", err)
		return
	}
	fmt.Printf("✓ Created wallet for address: %s\n", address)
}

func createJob(bc *blockchain.Blockchain, id, poster, title, description, payment, reputation string) {
	var paymentAmount, reputationReward int64
	fmt.Sscanf(payment, "%d", &paymentAmount)
	fmt.Sscanf(reputation, "%d", &reputationReward)

	tx := blockchain.Transaction{
		ID:   fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		Type: "job_create",
		From: poster,
		Data: map[string]interface{}{
			"job_id":            id,
			"title":             title,
			"description":       description,
			"payment":           float64(paymentAmount),
			"reputation_reward": float64(reputationReward),
		},
	}

	err := bc.AddTransaction(tx)
	if err != nil {
		fmt.Printf("Error creating job: %v\n", err)
		return
	}
	fmt.Printf("✓ Job created: %s\n", id)
}

func transfer(bc *blockchain.Blockchain, from, to, amount string) {
	var transferAmount int64
	fmt.Sscanf(amount, "%d", &transferAmount)

	tx := blockchain.Transaction{
		ID:     fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		Type:   "transfer",
		From:   from,
		To:     to,
		Amount: transferAmount,
	}

	err := bc.AddTransaction(tx)
	if err != nil {
		fmt.Printf("Error creating transfer: %v\n", err)
		return
	}
	fmt.Printf("✓ Transfer queued: %s -> %s (%d units)\n", from, to, transferAmount)
}

func mine(bc *blockchain.Blockchain) {
	err := bc.MineBlock()
	if err != nil {
		fmt.Printf("Error mining block: %v\n", err)
		return
	}
	fmt.Println("✓ Block mined successfully")
}

func getBalance(bc *blockchain.Blockchain, address string) {
	balance := bc.GetBalance(address)
	fmt.Printf("Balance for %s: %d units\n", address, balance)
}

func getReputation(bc *blockchain.Blockchain, address string) {
	rep, err := bc.GetReputation(address)
	if err != nil {
		fmt.Printf("Error getting reputation: %v\n", err)
		return
	}

	fmt.Printf("Reputation for %s: %d points\n", address, rep.Points)
	if len(rep.History) > 0 {
		fmt.Println("\nReputation History:")
		for _, event := range rep.History {
			fmt.Printf("  [%s] %+d points - %s\n", event.EventType, event.Points, event.Description)
		}
	}
}

func printChain(bc *blockchain.Blockchain) {
	chain := bc.GetChain()
	data, _ := json.MarshalIndent(chain, "", "  ")
	fmt.Println(string(data))
}

func runDemo(bc *blockchain.Blockchain) {
	fmt.Println("=== MeritChain Demo ===\n")

	// Create wallets
	fmt.Println("1. Creating wallets...")
	bc.CreateWallet("alice")
	bc.CreateWallet("bob")
	bc.CreateWallet("charlie")
	fmt.Println("✓ Created wallets for alice, bob, and charlie\n")

	// Simulate some initial balances (in real system, this would come from mining or initial distribution)
	fmt.Println("2. Setting initial balances (simulated)...")
	// Note: In production, this would require proper transactions
	fmt.Println("✓ Initial balances set\n")

	// Create a job
	fmt.Println("3. Alice creates a job...")
	tx1 := blockchain.Transaction{
		ID:   "tx-job-001",
		Type: "job_create",
		From: "alice",
		Data: map[string]interface{}{
			"job_id":            "job-001",
			"title":             "Build a website",
			"description":       "Need a professional website for my business",
			"payment":           float64(1000),
			"reputation_reward": float64(50),
		},
	}
	bc.AddTransaction(tx1)
	fmt.Println("✓ Job created: Build a website (1000 units, 50 reputation)\n")

	// Mine the block
	fmt.Println("4. Mining block...")
	bc.MineBlock()
	fmt.Println("✓ Block mined\n")

	// Bob accepts and completes the job
	fmt.Println("5. Bob accepts the job...")
	jobManager := bc.GetJobManager()
	jobManager.AcceptJob("job-001", "bob", time.Now().Unix())
	jobManager.StartJob("job-001", "bob")
	jobManager.SubmitJob("job-001", "bob")
	fmt.Println("✓ Job accepted and submitted by bob\n")

	// Add a review
	fmt.Println("6. Charlie reviews the work...")
	jobManager.AddReview("job-001", "charlie", 5, "Excellent work!", true, time.Now().Unix())
	fmt.Println("✓ Review added (5 stars, approved)\n")

	// Complete the job
	fmt.Println("7. Job completion transaction...")
	tx2 := blockchain.Transaction{
		ID:   "tx-complete-001",
		Type: "job_complete",
		From: "alice",
		To:   "bob",
		Data: map[string]interface{}{
			"job_id": "job-001",
		},
	}
	bc.AddTransaction(tx2)
	bc.MineBlock()
	fmt.Println("✓ Job completed, payment released, reputation awarded\n")

	// Show results
	fmt.Println("=== Results ===")
	fmt.Printf("Bob's balance: %d units\n", bc.GetBalance("bob"))
	
	if rep, err := bc.GetReputation("bob"); err == nil {
		fmt.Printf("Bob's reputation: %d points\n", rep.Points)
	}

	// Demonstrate fee scaling
	fmt.Println("\n=== Fee Scaling Demo ===")
	feeCalc := bc.GetFeeCalculator()
	
	fmt.Println("\nTransaction fees based on income:")
	incomes := []int64{0, 100, 1000, 10000, 100000}
	for _, income := range incomes {
		fee, _ := feeCalc.CalculateFee(income, 1000)
		fmt.Printf("  Income: %6d -> Fee: %4d units\n", income, fee)
	}

	fmt.Println("\nFees with reputation discount (50 reputation points):")
	for _, income := range incomes {
		fee, _ := feeCalc.CalculateFeeWithReputation(income, 1000, 50)
		fmt.Printf("  Income: %6d -> Fee: %4d units (with reputation discount)\n", income, fee)
	}

	// Show tokenomics stats
	fmt.Println("\n=== Tokenomics Stats ===")
	stats := bc.GetTokenomics().GetStats()
	fmt.Printf("Total Supply Cap: %d coins\n", stats.TotalSupplyCap)
	fmt.Printf("Emitted through Mining: %d coins (%.2f%% of genesis reserve)\n", stats.TotalMinted, stats.PercentMinted)
	fmt.Printf("Circulating Supply: %d coins\n", stats.CirculatingSupply)
	fmt.Printf("Current Block Height: %d\n", stats.CurrentBlockHeight)
	fmt.Printf("Current Block Reward: %d coins\n", stats.CurrentBlockReward)
	fmt.Printf("Next Halving at Block: %d\n", stats.NextHalvingBlock)
	fmt.Printf("\nGenesis Allocations:\n")
	fmt.Printf("  Reserve (locked for mining): %d coins (70%%)\n", stats.GenesisReserve)
	fmt.Printf("  Community Pool: %d coins (20%%)\n", stats.CommunityPool)
	fmt.Printf("  Protocol Reserve: %d coins (10%%)\n", stats.ProtocolReserve)
	fmt.Printf("\nReward Pools:\n")
	fmt.Printf("  Worker Reward Pool: %d coins\n", stats.WorkerRewardPool)
	fmt.Printf("  Reviewer Reward Pool: %d coins\n", stats.ReviewerRewardPool)

	fmt.Println("\n✓ Demo completed!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("  • 21 million coin hard cap with halving mechanism")
	fmt.Println("  • Genesis allocation: 70% mining reserve, 20% community, 10% protocol")
	fmt.Println("  • Block rewards split: 60% validators, 30% workers, 10% reviewers")
	fmt.Println("  • Non-transferable reputation earned through work")
	fmt.Println("  • Job escrow with peer review")
	fmt.Println("  • Income-based fee scaling (lower income = lower fees)")
	fmt.Println("  • Reputation discounts on transaction fees")
}

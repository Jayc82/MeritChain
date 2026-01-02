# MeritChain

A human-first blockchain where **reputation replaces identity**, **work replaces speculation**, and **access is based on contribution not wealth**.

## Overview

MeritChain is a blockchain system designed for dignity and opportunity. It features:

- **Non-transferable Reputation**: Reputation points earned through completed jobs and honest participation cannot be sold or transferred, ensuring authentic contribution tracking
- **Decentralized Job Contracts**: Job escrow system with peer review that enables trustless work agreements
- **Income-Based Transaction Fees**: Fees scale with on-chain income to keep the network accessible to everyone, regardless of economic status
- **Reputation Discounts**: Active, honest participants earn lower transaction fees as a reward

## Key Features

### 1. Reputation System

Reputation is earned through:
- Completing jobs successfully
- Positive peer reviews
- Honest participation in the network

Reputation cannot be:
- Transferred to other addresses
- Sold or bought
- Faked or manipulated

### 2. Job Escrow System

Jobs flow through a secure lifecycle:
1. **Created**: Poster creates job with payment locked in escrow
2. **Accepted**: Worker accepts the job
3. **In Progress**: Worker completes the work
4. **Submitted**: Worker submits for review
5. **Reviewed**: Peers review the work (1-5 stars, approve/dispute)
6. **Completed**: Payment released, reputation awarded

### 3. Income-Based Fee Scaling

Transaction fees are calculated based on the sender's on-chain income:
- **Low income (0-100 units)**: Minimum fee (10 units)
- **Medium income (1000 units)**: ~150 units
- **High income (100000 units)**: ~250 units

Fees use logarithmic scaling to ensure slow growth with income, keeping the network usable for everyone.

### 4. Reputation Discounts

Users with higher reputation earn discounts on transaction fees (up to 50% off), incentivizing honest participation and quality work.

## Installation

### Prerequisites

- Go 1.21 or higher

### Build

```bash
go build -o meritchain ./cmd/meritchain
```

## Usage

### Run the Demo

See MeritChain in action with a complete demonstration:

```bash
./meritchain demo
```

This demonstrates:
- Creating wallets with reputation tracking
- Job creation with escrow
- Job acceptance and completion workflow
- Peer review system
- Income-based fee calculation
- Reputation-based fee discounts

### CLI Commands

#### Create a wallet
```bash
./meritchain wallet <address>
```

Example:
```bash
./meritchain wallet alice
```

#### Create a job
```bash
./meritchain job <id> <poster> <title> <description> <payment> <reputation>
```

Example:
```bash
./meritchain job job-001 alice "Build website" "Need a professional site" 1000 50
```

#### Transfer funds
```bash
./meritchain transfer <from> <to> <amount>
```

Example:
```bash
./meritchain transfer alice bob 500
```

#### Mine pending transactions
```bash
./meritchain mine
```

#### Check balance
```bash
./meritchain balance <address>
```

Example:
```bash
./meritchain balance alice
```

#### Check reputation
```bash
./meritchain reputation <address>
```

Example:
```bash
./meritchain reputation bob
```

#### View blockchain
```bash
./meritchain chain
```

## Architecture

### Core Packages

#### `pkg/reputation`
Manages non-transferable reputation points and history for each wallet. Tracks how reputation was earned through job completion and peer reviews.

#### `pkg/job`
Handles job contracts with escrow, including:
- Job creation and lifecycle management
- Worker assignment and submission
- Peer review system with ratings
- Escrow release upon completion

#### `pkg/fee`
Calculates transaction fees based on:
- Sender's on-chain income (logarithmic scaling)
- Reputation level (up to 50% discount)
- Configurable parameters (base fee, min/max, scale factor)

#### `pkg/blockchain`
Core blockchain implementation with:
- Block and transaction structures
- Block mining and validation
- Integration of reputation, job, and fee systems
- State management for balances and contracts

## Testing

Run all tests:

```bash
go test ./pkg/...
```

Run tests for a specific package:

```bash
go test ./pkg/reputation -v
go test ./pkg/job -v
go test ./pkg/fee -v
```

## Design Philosophy

MeritChain is built on the principle that blockchain should serve humanity, not speculation. The design choices reflect this:

1. **Reputation over Identity**: Your contribution history matters more than who you are
2. **Work over Speculation**: Value comes from completing jobs, not trading tokens
3. **Accessibility over Exclusivity**: Fee scaling ensures the network remains usable for everyone
4. **Merit over Wealth**: Higher reputation (earned through work) provides tangible benefits

## Future Enhancements

Potential areas for expansion:
- Dispute resolution system for job conflicts
- Multi-signature job approval
- Reputation decay for inactive accounts
- Cross-job reputation dependencies
- Stake-based consensus mechanism
- Web interface for job marketplace
- Mobile wallet application

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.


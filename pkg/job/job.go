package job

import (
	"errors"
	"sync"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	JobStatusOpen       JobStatus = "open"
	JobStatusAccepted   JobStatus = "accepted"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusSubmitted  JobStatus = "submitted"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusDisputed   JobStatus = "disputed"
	JobStatusCancelled  JobStatus = "cancelled"
)

// Job represents a decentralized job contract with escrow
type Job struct {
	ID              string
	Poster          string  // Address of job poster
	Worker          string  // Address of worker (empty until accepted)
	Title           string
	Description     string
	Payment         int64   // Payment amount in escrow
	ReputationReward int64  // Reputation points on completion
	Status          JobStatus
	CreatedAt       int64
	AcceptedAt      int64
	CompletedAt     int64
	Reviews         []Review
}

// Review represents a peer review for job completion
type Review struct {
	Reviewer    string
	Rating      int    // 1-5 stars
	Comment     string
	Timestamp   int64
	IsApproved  bool   // true = approve completion, false = dispute
}

// JobManager manages all job contracts
type JobManager struct {
	mu    sync.RWMutex
	jobs  map[string]*Job
	escrow map[string]int64 // jobID -> escrowed amount
}

// NewJobManager creates a new job manager
func NewJobManager() *JobManager {
	return &JobManager{
		jobs:   make(map[string]*Job),
		escrow: make(map[string]int64),
	}
}

// CreateJob creates a new job with escrow
func (jm *JobManager) CreateJob(id, poster, title, description string, payment, reputationReward, timestamp int64) (*Job, error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if _, exists := jm.jobs[id]; exists {
		return nil, errors.New("job ID already exists")
	}

	job := &Job{
		ID:               id,
		Poster:           poster,
		Worker:           "",
		Title:            title,
		Description:      description,
		Payment:          payment,
		ReputationReward: reputationReward,
		Status:           JobStatusOpen,
		CreatedAt:        timestamp,
		Reviews:          []Review{},
	}

	jm.jobs[id] = job
	jm.escrow[id] = payment

	return job, nil
}

// AcceptJob allows a worker to accept an open job
func (jm *JobManager) AcceptJob(jobID, worker string, timestamp int64) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return errors.New("job not found")
	}

	if job.Status != JobStatusOpen {
		return errors.New("job is not open for acceptance")
	}

	job.Worker = worker
	job.Status = JobStatusAccepted
	job.AcceptedAt = timestamp

	return nil
}

// StartJob marks a job as in progress
func (jm *JobManager) StartJob(jobID, worker string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return errors.New("job not found")
	}

	if job.Worker != worker {
		return errors.New("only the assigned worker can start the job")
	}

	if job.Status != JobStatusAccepted {
		return errors.New("job must be accepted before starting")
	}

	job.Status = JobStatusInProgress
	return nil
}

// SubmitJob submits completed work for review
func (jm *JobManager) SubmitJob(jobID, worker string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return errors.New("job not found")
	}

	if job.Worker != worker {
		return errors.New("only the assigned worker can submit the job")
	}

	if job.Status != JobStatusInProgress {
		return errors.New("job must be in progress to submit")
	}

	job.Status = JobStatusSubmitted
	return nil
}

// AddReview adds a peer review to a submitted job
func (jm *JobManager) AddReview(jobID, reviewer string, rating int, comment string, isApproved bool, timestamp int64) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return errors.New("job not found")
	}

	if job.Status != JobStatusSubmitted {
		return errors.New("job must be submitted for review")
	}

	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	review := Review{
		Reviewer:   reviewer,
		Rating:     rating,
		Comment:    comment,
		Timestamp:  timestamp,
		IsApproved: isApproved,
	}

	job.Reviews = append(job.Reviews, review)
	return nil
}

// CompleteJob completes a job and releases escrow (requires poster approval or positive reviews)
func (jm *JobManager) CompleteJob(jobID string, timestamp int64) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return errors.New("job not found")
	}

	if job.Status != JobStatusSubmitted {
		return errors.New("job must be submitted before completion")
	}

	// Check if there are any approving reviews
	hasApproval := false
	for _, review := range job.Reviews {
		if review.IsApproved {
			hasApproval = true
			break
		}
	}

	if !hasApproval && len(job.Reviews) > 0 {
		return errors.New("job requires approval from reviews")
	}

	job.Status = JobStatusCompleted
	job.CompletedAt = timestamp

	// Release escrow (in a real system, this would transfer payment)
	delete(jm.escrow, jobID)

	return nil
}

// GetJob returns a job by ID
func (jm *JobManager) GetJob(jobID string) (*Job, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return nil, errors.New("job not found")
	}

	return job, nil
}

// GetAllJobs returns all jobs
func (jm *JobManager) GetAllJobs() []*Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	jobs := make([]*Job, 0, len(jm.jobs))
	for _, job := range jm.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// GetEscrowAmount returns the escrowed amount for a job
func (jm *JobManager) GetEscrowAmount(jobID string) (int64, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	amount, exists := jm.escrow[jobID]
	if !exists {
		return 0, errors.New("no escrow found for job")
	}
	return amount, nil
}

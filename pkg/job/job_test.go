package job

import (
	"testing"
)

func TestJobManager_CreateJob(t *testing.T) {
	jm := NewJobManager()
	
	job, err := jm.CreateJob("job-001", "alice", "Build website", "Need a website", 1000, 50, 1234567890)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}
	
	if job.ID != "job-001" {
		t.Errorf("Expected job ID 'job-001', got '%s'", job.ID)
	}
	
	if job.Status != JobStatusOpen {
		t.Errorf("Expected status 'open', got '%s'", job.Status)
	}
	
	if job.Payment != 1000 {
		t.Errorf("Expected payment 1000, got %d", job.Payment)
	}
	
	escrow, err := jm.GetEscrowAmount("job-001")
	if err != nil {
		t.Fatalf("Failed to get escrow: %v", err)
	}
	
	if escrow != 1000 {
		t.Errorf("Expected escrow 1000, got %d", escrow)
	}
}

func TestJobManager_AcceptJob(t *testing.T) {
	jm := NewJobManager()
	jm.CreateJob("job-002", "alice", "Design logo", "Need a logo", 500, 25, 1234567890)
	
	err := jm.AcceptJob("job-002", "bob", 1234567891)
	if err != nil {
		t.Fatalf("Failed to accept job: %v", err)
	}
	
	job, _ := jm.GetJob("job-002")
	if job.Worker != "bob" {
		t.Errorf("Expected worker 'bob', got '%s'", job.Worker)
	}
	
	if job.Status != JobStatusAccepted {
		t.Errorf("Expected status 'accepted', got '%s'", job.Status)
	}
}

func TestJobManager_JobWorkflow(t *testing.T) {
	jm := NewJobManager()
	jm.CreateJob("job-003", "alice", "Write article", "Need content", 300, 15, 1234567890)
	
	// Accept job
	err := jm.AcceptJob("job-003", "charlie", 1234567891)
	if err != nil {
		t.Fatalf("Failed to accept job: %v", err)
	}
	
	// Start job
	err = jm.StartJob("job-003", "charlie")
	if err != nil {
		t.Fatalf("Failed to start job: %v", err)
	}
	
	job, _ := jm.GetJob("job-003")
	if job.Status != JobStatusInProgress {
		t.Errorf("Expected status 'in_progress', got '%s'", job.Status)
	}
	
	// Submit job
	err = jm.SubmitJob("job-003", "charlie")
	if err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}
	
	job, _ = jm.GetJob("job-003")
	if job.Status != JobStatusSubmitted {
		t.Errorf("Expected status 'submitted', got '%s'", job.Status)
	}
}

func TestJobManager_AddReview(t *testing.T) {
	jm := NewJobManager()
	jm.CreateJob("job-004", "alice", "Code review", "Review code", 200, 10, 1234567890)
	jm.AcceptJob("job-004", "bob", 1234567891)
	jm.StartJob("job-004", "bob")
	jm.SubmitJob("job-004", "bob")
	
	err := jm.AddReview("job-004", "reviewer1", 5, "Excellent work", true, 1234567892)
	if err != nil {
		t.Fatalf("Failed to add review: %v", err)
	}
	
	job, _ := jm.GetJob("job-004")
	if len(job.Reviews) != 1 {
		t.Errorf("Expected 1 review, got %d", len(job.Reviews))
	}
	
	if job.Reviews[0].Rating != 5 {
		t.Errorf("Expected rating 5, got %d", job.Reviews[0].Rating)
	}
	
	if !job.Reviews[0].IsApproved {
		t.Error("Expected review to be approved")
	}
}

func TestJobManager_CompleteJob(t *testing.T) {
	jm := NewJobManager()
	jm.CreateJob("job-005", "alice", "Test task", "Testing", 100, 5, 1234567890)
	jm.AcceptJob("job-005", "bob", 1234567891)
	jm.StartJob("job-005", "bob")
	jm.SubmitJob("job-005", "bob")
	jm.AddReview("job-005", "reviewer1", 4, "Good", true, 1234567892)
	
	err := jm.CompleteJob("job-005", 1234567893)
	if err != nil {
		t.Fatalf("Failed to complete job: %v", err)
	}
	
	job, _ := jm.GetJob("job-005")
	if job.Status != JobStatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", job.Status)
	}
	
	// Escrow should be released
	_, err = jm.GetEscrowAmount("job-005")
	if err == nil {
		t.Error("Expected escrow to be released")
	}
}

func TestJobManager_InvalidRating(t *testing.T) {
	jm := NewJobManager()
	jm.CreateJob("job-006", "alice", "Task", "Description", 100, 5, 1234567890)
	jm.AcceptJob("job-006", "bob", 1234567891)
	jm.StartJob("job-006", "bob")
	jm.SubmitJob("job-006", "bob")
	
	err := jm.AddReview("job-006", "reviewer1", 6, "Too high", true, 1234567892)
	if err == nil {
		t.Error("Expected error for invalid rating")
	}
	
	err = jm.AddReview("job-006", "reviewer1", 0, "Too low", true, 1234567892)
	if err == nil {
		t.Error("Expected error for invalid rating")
	}
}

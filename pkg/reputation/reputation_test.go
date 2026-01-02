package reputation

import (
	"testing"
)

func TestReputationManager_CreateWallet(t *testing.T) {
	rm := NewReputationManager()
	
	err := rm.CreateWallet("alice")
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}
	
	rep, err := rm.GetReputation("alice")
	if err != nil {
		t.Fatalf("Failed to get reputation: %v", err)
	}
	
	if rep.Address != "alice" {
		t.Errorf("Expected address 'alice', got '%s'", rep.Address)
	}
	
	if rep.Points != 0 {
		t.Errorf("Expected 0 initial points, got %d", rep.Points)
	}
}

func TestReputationManager_AddReputation(t *testing.T) {
	rm := NewReputationManager()
	rm.CreateWallet("bob")
	
	err := rm.AddReputation("bob", 50, "job_completed", "Completed a job", "job-001", 1234567890)
	if err != nil {
		t.Fatalf("Failed to add reputation: %v", err)
	}
	
	rep, _ := rm.GetReputation("bob")
	if rep.Points != 50 {
		t.Errorf("Expected 50 points, got %d", rep.Points)
	}
	
	if len(rep.History) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(rep.History))
	}
	
	if rep.History[0].EventType != "job_completed" {
		t.Errorf("Expected event type 'job_completed', got '%s'", rep.History[0].EventType)
	}
}

func TestReputationManager_RemoveReputation(t *testing.T) {
	rm := NewReputationManager()
	rm.CreateWallet("charlie")
	rm.AddReputation("charlie", 100, "job_completed", "Initial reputation", "job-001", 1234567890)
	
	err := rm.RemoveReputation("charlie", 30, "peer_review_negative", "Negative review", "review-001", 1234567891)
	if err != nil {
		t.Fatalf("Failed to remove reputation: %v", err)
	}
	
	rep, _ := rm.GetReputation("charlie")
	if rep.Points != 70 {
		t.Errorf("Expected 70 points, got %d", rep.Points)
	}
}

func TestReputationManager_ReputationCannotGoNegative(t *testing.T) {
	rm := NewReputationManager()
	rm.CreateWallet("dave")
	
	rm.RemoveReputation("dave", 100, "penalty", "Large penalty", "penalty-001", 1234567890)
	
	rep, _ := rm.GetReputation("dave")
	if rep.Points != 0 {
		t.Errorf("Expected 0 points (floor), got %d", rep.Points)
	}
}

func TestReputationManager_DuplicateWallet(t *testing.T) {
	rm := NewReputationManager()
	rm.CreateWallet("eve")
	
	err := rm.CreateWallet("eve")
	if err == nil {
		t.Error("Expected error when creating duplicate wallet")
	}
}

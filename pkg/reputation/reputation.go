package reputation

import (
	"errors"
	"sync"
)

// Reputation represents non-transferable reputation points earned through work
type Reputation struct {
	Address string
	Points  int64
	History []ReputationEvent
}

// ReputationEvent tracks how reputation was earned or lost
type ReputationEvent struct {
	Timestamp   int64
	EventType   string // "job_completed", "peer_review_positive", "peer_review_negative"
	Points      int64
	Description string
	RelatedID   string // Job ID or review ID
}

// ReputationManager manages all reputation accounts
type ReputationManager struct {
	mu          sync.RWMutex
	reputations map[string]*Reputation
}

// NewReputationManager creates a new reputation manager
func NewReputationManager() *ReputationManager {
	return &ReputationManager{
		reputations: make(map[string]*Reputation),
	}
}

// CreateWallet creates a new reputation wallet for an address
func (rm *ReputationManager) CreateWallet(address string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.reputations[address]; exists {
		return errors.New("wallet already exists")
	}

	rm.reputations[address] = &Reputation{
		Address: address,
		Points:  0,
		History: []ReputationEvent{},
	}
	return nil
}

// GetReputation returns the reputation for an address
func (rm *ReputationManager) GetReputation(address string) (*Reputation, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rep, exists := rm.reputations[address]
	if !exists {
		return nil, errors.New("wallet not found")
	}
	return rep, nil
}

// AddReputation adds reputation points for an event
func (rm *ReputationManager) AddReputation(address string, points int64, eventType, description, relatedID string, timestamp int64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rep, exists := rm.reputations[address]
	if !exists {
		return errors.New("wallet not found")
	}

	rep.Points += points
	rep.History = append(rep.History, ReputationEvent{
		Timestamp:   timestamp,
		EventType:   eventType,
		Points:      points,
		Description: description,
		RelatedID:   relatedID,
	})

	return nil
}

// RemoveReputation removes reputation points (for negative actions)
func (rm *ReputationManager) RemoveReputation(address string, points int64, eventType, description, relatedID string, timestamp int64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rep, exists := rm.reputations[address]
	if !exists {
		return errors.New("wallet not found")
	}

	rep.Points -= points
	if rep.Points < 0 {
		rep.Points = 0 // Reputation can't go negative
	}

	rep.History = append(rep.History, ReputationEvent{
		Timestamp:   timestamp,
		EventType:   eventType,
		Points:      -points,
		Description: description,
		RelatedID:   relatedID,
	})

	return nil
}

// GetAllReputations returns all reputation accounts
func (rm *ReputationManager) GetAllReputations() map[string]*Reputation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]*Reputation)
	for k, v := range rm.reputations {
		result[k] = v
	}
	return result
}

package db

import (
	"database/sql"
	"time"
)

type CircuitBreaker struct {
	ID               string `json:"id"`
	Provider         string `json:"provider"`
	State            string `json:"state"` // closed, open, half-open
	FailureCount     int    `json:"failureCount"`
	LastFailureAt    string `json:"lastFailureAt,omitempty"`
	LastStateChangeAt string `json:"lastStateChangeAt"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

func GetCircuitBreaker(db *sql.DB, provider string) (*CircuitBreaker, error) {
	var cb CircuitBreaker
	err := db.QueryRow(
		"SELECT id, provider, state, failure_count, COALESCE(last_failure_at,''), last_state_change_at, created_at, updated_at FROM domain_circuit_breakers WHERE provider = ?",
		provider).Scan(&cb.ID, &cb.Provider, &cb.State, &cb.FailureCount, &cb.LastFailureAt, &cb.LastStateChangeAt, &cb.CreatedAt, &cb.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cb, nil
}

func UpsertCircuitBreaker(db *sql.DB, cb CircuitBreaker) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(
		"INSERT INTO domain_circuit_breakers (id, provider, state, failure_count, last_failure_at, last_state_change_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?) "+
			"ON CONFLICT(id) DO UPDATE SET state=excluded.state, failure_count=excluded.failure_count, last_failure_at=excluded.last_failure_at, last_state_change_at=excluded.last_state_change_at, updated_at=excluded.updated_at",
		cb.ID, cb.Provider, cb.State, cb.FailureCount, cb.LastFailureAt, now, now)
	return err
}

func RecordProviderFailure(db *sql.DB, provider string) {
	cb, err := GetCircuitBreaker(db, provider)
	if err != nil || cb == nil {
		// Create new circuit breaker
		cb = &CircuitBreaker{
			ID:               provider,
			Provider:         provider,
			State:            "closed",
			FailureCount:     1,
			LastFailureAt:    time.Now().UTC().Format(time.RFC3339),
			LastStateChangeAt: time.Now().UTC().Format(time.RFC3339),
		}
	} else {
		cb.FailureCount++
		cb.LastFailureAt = time.Now().UTC().Format(time.RFC3339)
		// Open circuit after 5 consecutive failures
		if cb.FailureCount >= 5 && cb.State != "open" {
			cb.State = "open"
			cb.LastStateChangeAt = time.Now().UTC().Format(time.RFC3339)
		}
	}
	UpsertCircuitBreaker(db, *cb)
}

func RecordProviderSuccess(db *sql.DB, provider string) {
	cb, err := GetCircuitBreaker(db, provider)
	if err != nil || cb == nil {
		return
	}
	cb.FailureCount = 0
	cb.State = "closed"
	cb.LastStateChangeAt = time.Now().UTC().Format(time.RFC3339)
	UpsertCircuitBreaker(db, *cb)
}

func IsProviderCircuitOpen(db *sql.DB, provider string) bool {
	cb, err := GetCircuitBreaker(db, provider)
	if err != nil || cb == nil {
		return false
	}
	return cb.State == "open"
}

package memory

import (
	"context"
	"sync"
)

// Memory defines the interface for agents' memory system
type Memory interface {
	// Store saves a record to memory
	Store(ctx context.Context, record Record) error

	// Retrieve fetches records from memory based on a query
	Retrieve(ctx context.Context, query string, limit int) ([]Record, error)

	// Clear removes all records from memory
	Clear(ctx context.Context) error
}

// Record represents a single memory entry
type Record struct {
	AgentName string
	Content   string
	Timestamp int64 // Unix timestamp (optional for implementations)
}

// Base implements basic memory storage in-memory
// Note: For production use, consider implementing persistent storage
type Base struct {
	mu      sync.RWMutex
	records []Record
}

// New creates a new Base memory storage
func New() *Base {
	return &Base{
		records: make([]Record, 0),
	}
}

// Store saves a record to memory
func (m *Base) Store(ctx context.Context, record Record) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.mu.Lock()
		m.records = append(m.records, record)
		m.mu.Unlock()
		return nil
	}
}

// Retrieve fetches records from memory based on a simple keyword match
// This is a basic implementation - in production, use vector similarity search
func (m *Base) Retrieve(ctx context.Context, query string, limit int) ([]Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.records) {
		limit = len(m.records)
	}

	records := make([]Record, 0, limit)

	// For now, just return the most recent records
	startIdx := len(m.records) - limit
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(m.records); i++ {
		records = append(records, m.records[i])
	}

	return records, nil
}

// Clear removes all records from memory
func (m *Base) Clear(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.mu.Lock()
		m.records = make([]Record, 0)
		m.mu.Unlock()
		return nil
	}
}

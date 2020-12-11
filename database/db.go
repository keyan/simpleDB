package database

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// memStorage is the representation of in-memory data.
// A new type is declared because this is shared with the journal.
type memStorage map[string][]byte

// Database is the main interface for data storage and access for callers.
type Database struct {
	// The underlying data held by the DB.
	data memStorage

	// The write-ahead-log used to track changes and checkpoint full state.
	j journal

	// A table of locks for each key in the underlying datastore.
	// Concurrent access is allowed so each key must be protected from reads during
	// a write operation.
	locks map[string]sync.RWMutex

	// Central lock must be acquired when adding a new key to `data` because that
	// requires also adding a new RWMutex to `locks`.
	locksMutex sync.RWMutex
}

// New creates a new Database struct, loads any available data from disk, schedules
// regular checkpointing, then returns a pointer to the Database indicating it is ready
// to serve requests.
func New() *Database {
	db := &Database{
		data:  map[string][]byte{},
		locks: map[string]sync.RWMutex{},
		j:     newJournal(),
	}

	err := db.loadStateFromDisk()
	if err != nil {
		panic(fmt.Sprintf(
			"Could not load DB state from disk during startup, %v", err))
	}

	go db.scheduleCheckpoints()

	return db
}

// Get stores returns the value for provided key. If not present in the Database then
// an error is returned instead.
func (s *Database) Get(key string) ([]byte, error) {
	s.locksMutex.RLock()
	mu, ok := s.locks[key]
	s.locksMutex.RUnlock()
	if !ok {
		return nil, errors.New("Key not found")
	}
	mu.RLock()
	defer mu.RUnlock()

	val, ok := s.data[key]
	if !ok {
		return nil, errors.New("Key not found, but lock present")
	}

	return val, nil
}

// Set updates the value for provided key. If not already present the value is silently added.
func (s *Database) Set(key string, value []byte) {
	s.locksMutex.RLock()
	mu, ok := s.locks[key]
	s.locksMutex.RUnlock()

	// This key has never been set, so it needs a new lock too.
	if !ok {
		s.locksMutex.Lock()
		mu = sync.RWMutex{}
		s.locks[key] = mu
		s.locksMutex.Unlock()
	}
	mu.Lock()
	defer mu.Unlock()

	s.j.addWriteOp(key, value)
	s.data[key] = value
}

// Delete removes the value for the provided key.
func (s *Database) Delete(key string) {
	s.locksMutex.RLock()
	mu, ok := s.locks[key]
	s.locksMutex.RUnlock()
	if !ok {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	s.j.addRemoveOp(key)
	delete(s.data, key)
}

// loadStateFromDisk is run when a new Database{} is created so that any on-disk state is
// reloaded into the Database before providing callers access. If no prior on-disk data is
// available the database is simply empty at creation.
func (s *Database) loadStateFromDisk() error {
	return s.j.load(s.data)
}

// scheduleCheckpoints captures the entire db state to disk after some amount of time. It should
// be run in a separate goroutine. How often the db is checkpointed is a matter of whether to trade
// off startup time or runtime performance, i.e. more checkpoints will slow down runtime performance.
func (s *Database) scheduleCheckpoints() error {
	for {
		time.Sleep(10 * time.Second)

		s.locksMutex.Lock()
		s.j.checkpoint(s.data)
		s.locksMutex.Unlock()
	}
}

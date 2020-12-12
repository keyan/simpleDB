package database

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/keyan/simpledb/rpc"
)

// memStorage is the representation of in-memory data.
// A new type is declared because this is shared with the journal.
type memStorage map[string]rpc.ValueType

// Database is the main interface for data storage and access for callers.
type Database struct {
	// The underlying data held by the DB.
	data memStorage

	// The write-ahead-log used to track changes and checkpoint full state.
	j journal

	// Central lock is required for Set/Delete operations. Could be possible
	// to do key-level locking which would allow update Set operations to not
	// take a global lock, but not doing that for now.
	sync.RWMutex
}

// New creates a new Database struct, loads any available data from disk, schedules
// regular checkpointing, then returns a pointer to the Database indicating it is ready
// to serve requests.
func New() *Database {
	db := &Database{
		data: memStorage{},
		j:    newJournal(),
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
func (s *Database) Get(key string) (rpc.ValueType, error) {
	s.RLock()
	defer s.RUnlock()

	val, ok := s.data[key]
	if !ok {
		return nil, errors.New("Key not found")
	}

	return val, nil
}

// Set updates the value for provided key. If not already present the value is silently added.
func (s *Database) Set(key string, value rpc.ValueType) {
	s.Lock()
	defer s.Unlock()

	s.j.addWriteOp(key, value)
	s.data[key] = value
}

// Delete removes the value for the provided key.
func (s *Database) Delete(key string) {
	s.Lock()
	defer s.Unlock()

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

		s.Lock()
		s.j.checkpoint(s.data)
		s.Unlock()
	}
}

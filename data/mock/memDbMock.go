package mock

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
)

// MemDbMock represents the memory database storage. It holds a map of key value pairs
// and a mutex to handle concurrent accesses to the map
type MemDbMock struct {
	db   map[string][]byte
	mutx sync.RWMutex
}

// NewMemDbMock creates a new memorydb object
func NewMemDbMock() (*MemDbMock, error) {
	return &MemDbMock{
		db:   make(map[string][]byte),
		mutx: sync.RWMutex{},
	}, nil
}

// Put adds the value to the (key, val) storage medium
func (s *MemDbMock) Put(key, val []byte) error {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	s.db[string(key)] = val

	return nil
}

// Get gets the value associated to the key, or reports an error
func (s *MemDbMock) Get(key []byte) ([]byte, error) {
	s.mutx.RLock()
	defer s.mutx.RUnlock()

	val, ok := s.db[string(key)]

	if !ok {
		return nil, errors.New(fmt.Sprintf("key: %s not found", base64.StdEncoding.EncodeToString(key)))
	}

	return val, nil
}

// Has returns true if the given key is present in the persistance medium, false otherwise
func (s *MemDbMock) Has(key []byte) (bool, error) {
	s.mutx.RLock()
	defer s.mutx.RUnlock()

	_, ok := s.db[string(key)]

	return ok, nil
}

// Init initializes the storage medium and prepares it for usage
func (s *MemDbMock) Init() error {
	// no special initialization needed
	return nil
}

// Close closes the files/resources associated to the storage medium
func (s *MemDbMock) Close() error {
	// nothing to do
	return nil
}

// Remove removes the data associated to the given key
func (s *MemDbMock) Remove(key []byte) error {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	delete(s.db, string(key))

	return nil
}

// Destroy removes the storage medium stored data
func (s *MemDbMock) Destroy() error {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	s.db = make(map[string][]byte)

	return nil
}

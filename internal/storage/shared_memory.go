package storage

import "sync"

// SharedMemory is shared across all pages.
type SharedMemory interface {
	// Get the value for a given key.
	Get(key string) (value any, err error)

	// Set the value for a given key.
	Set(key string, value any) (err error)

	// Delete the value for a given key.
	Delete(key string) (err error)

	// List the keys.
	List() (keys []string, err error)

	// Clear the storage.
	Clear() (err error)
}

type SharedMemoryImpl struct {
	data map[string]any
	mu   sync.RWMutex
}

func NewSharedMemory() SharedMemory {
	return &SharedMemoryImpl{
		data: make(map[string]any),
	}
}

func (s *SharedMemoryImpl) Get(key string) (value any, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key], nil
}

func (s *SharedMemoryImpl) Set(key string, value any) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *SharedMemoryImpl) Delete(key string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *SharedMemoryImpl) List() (keys []string, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys = make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}
	return keys, nil
}

func (s *SharedMemoryImpl) Clear() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]any)
	return nil
}

package store

import "sync"

type StreamStore struct {
	store map[string]map[string]string
	lock  sync.RWMutex
}

var StreamStoreInstance = &StreamStore{
	store: make(map[string]map[string]string),
}

func (s *StreamStore) AddItem(key string, data map[string]string) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store[key] = data

	return data["id"]

}

func (s *StreamStore) ContainsKey(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.store[key]
	return ok
}

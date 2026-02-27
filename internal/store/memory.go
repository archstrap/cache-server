package store

import (
	"strconv"
	"strings"
	"sync"
)

type StreamStore struct {
	store map[string][]map[string]string
	lock  sync.RWMutex
}

var StreamStoreInstance = &StreamStore{
	store: make(map[string][]map[string]string),
}

func (s *StreamStore) AddItem(key string, data map[string]string) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store[key] = append(s.store[key], data)

	return data["id"]

}

func (s *StreamStore) ContainsKey(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.containsKey(key)
}

func (s *StreamStore) containsKey(key string) bool {
	_, ok := s.store[key]
	return ok
}

func (s *StreamStore) IsValid(key string, data map[string]string) bool {

	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.containsKey(key) {
		return true
	}

	entries := s.store[key] // []map[string]string
	n := len(entries)
	lastEntry := entries[n-1]
	lastEntryId := lastEntry["id"]
	currentId := data["id"]

	return compareId(lastEntryId, currentId)
}

func compareId(id1 string, id2 string) bool {

	if !strings.Contains(id1, "-") || !strings.Contains(id2, "-") {
		return false
	}

	ok1, t1, s1 := extractIdDetails(id1)
	ok2, t2, s2 := extractIdDetails(id2)

	if !ok1 || !ok2 {
		return false
	}

	return t1 <= t2 && s1 < s2
}

func extractIdDetails(id string) (bool, int64, int) {
	s := strings.Split(id, "-")

	timeStamp, err := strconv.Atoi(s[0])
	if err != nil {
		return false, -1, -1
	}

	sequenceNo, err := strconv.Atoi(s[1])
	if err != nil {
		return false, -1, -1
	}

	return true, int64(timeStamp), sequenceNo
}

package store

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StreamStore struct {
	store map[string][]map[string]string
	lock  sync.RWMutex
}

var StreamStoreInstance = &StreamStore{
	store: make(map[string][]map[string]string),
}

func (s *StreamStore) ValidateAndAdd(key string, data map[string]string) (bool, string) {

	s.lock.Lock()
	defer s.lock.Unlock()

	data["id"] = s.generateId(data["id"], key)

	isValid := s.IsValid(key, data)
	if !isValid {
		return false, ""
	}

	return true, s.AddItem(key, data)
}

func (s *StreamStore) AddItem(key string, data map[string]string) string {

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

	if t2 > t1 {
		return true
	}

	return t1 == t2 && s1 < s2
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

func (s *StreamStore) generateId(id string, key string) string {
	if !strings.Contains(id, "*") {
		return id
	}

	var timeStamp string
	if "*" == id { // fully auto generated
		timeStamp = fmt.Sprintf("%d", time.Now().Local().UnixMilli())
	} else { // partially auto generated
		timeStamp = strings.Split(id, "-")[0]
	}

	if !s.containsKey(key) {
		id := fmt.Sprintf("%s-%d", timeStamp, 0)
		if id == "0-0" {
			return "0-1"
		}
		return id
	}

	entries := s.store[key]
	lastEntry := entries[len(entries)-1]
	lastEntryId := lastEntry["id"]
	lastEntryArgs := strings.Split(lastEntryId, "-")
	lastEntryTimeStamp := lastEntryArgs[0]

	if timeStamp != lastEntryTimeStamp {
		id := fmt.Sprintf("%s-%d", timeStamp, 0)
		if id == "0-0" {
			return "0-1"
		}

		return id
	}

	lastEntrySeqNo := lastEntryArgs[1]
	seqNo, _ := strconv.Atoi(lastEntrySeqNo)

	return fmt.Sprintf("%s-%d", timeStamp, (seqNo + 1))
}

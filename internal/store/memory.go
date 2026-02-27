package store

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StreamStore struct {
	store     map[string][]map[string]string
	timeStore map[string]time.Time
	lock      sync.RWMutex
	cond      *sync.Cond
}

var StreamStoreInstance = &StreamStore{
	store:     make(map[string][]map[string]string),
	timeStore: make(map[string]time.Time),
}

func init() {
	StreamStoreInstance.cond = sync.NewCond(StreamStoreInstance.lock.RLocker())
}

type Pair[K, V any] struct {
	k K
	v V
}

func NewPair[K, V any](k K, v V) *Pair[K, V] {
	return &Pair[K, V]{
		k: k,
		v: v,
	}
}

func (p *Pair[K, V]) GetK() K {
	return p.k
}

func (p *Pair[K, V]) GetV() V {
	return p.v
}

func (s *StreamStore) ValidateAndAdd(key string, data map[string]string) (bool, string) {

	s.lock.Lock()
	defer s.lock.Unlock()

	data["id"] = s.generateId(data["id"], key)

	isValid := s.IsValid(key, data)
	if !isValid {
		return false, ""
	}

	insertedId := s.AddItem(key, data)
	s.cond.Broadcast()
	return true, insertedId
}

func (s *StreamStore) AddItem(key string, data map[string]string) string {

	id := data["id"]
	s.store[key] = append(s.store[key], data)
	s.timeStore[id] = time.Now()

	return id

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

func (s *StreamStore) SearchInRange(key, start, end string) []any {
	s.lock.RLock()
	defer s.lock.RUnlock()

	entries := s.store[key]

	l := 0
	r := len(entries) - 1

	if start != "-" {
		for i := 0; i <= r; i++ {
			currentId := entries[i]["id"]
			if compare(currentId, start, greaterOrEqualTo) {
				l = i
				break
			}
		}
	}

	if end != "+" {
		for i := r; i >= l; i-- {
			currentId := entries[i]["id"]
			if compare(currentId, end, lesserOrEqualTo) {
				r = i
				break
			}
		}
	}

	result := make([]any, 0)

	for l <= r {
		currentEntry := entries[l]
		currentEntryId := currentEntry["id"]

		items := make([]string, 0)
		for k, v := range currentEntry {
			if k == "id" {
				continue
			}
			items = append(items, k, v)
		}

		currentResult := []any{
			currentEntryId,
			items,
		}

		result = append(result, currentResult)
		l++

	}

	return result
}

func (s *StreamStore) SearchExclusiveWithoutBlock(items []*Pair[string, string]) []any {

	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.SearchExclusiveAll(items)
}

func (s *StreamStore) SearchExclusiveWithBlock(items []*Pair[string, string], timeOut int, isBlockedForNewKey bool, startTime time.Time) []any {

	s.lock.RLock()
	defer s.lock.RUnlock()

	var deadline time.Time

	if timeOut != 0 {
		deadline = time.Now().Add(time.Duration(timeOut) * time.Millisecond)
	}

	// in case nothing gets received we will call up goroutine
	go func() {
		time.Sleep(time.Duration(timeOut) * time.Millisecond)
		s.cond.Broadcast()
	}()

	for {
		var result []any
		if isBlockedForNewKey {
			result = s.SearchExclusiveAllNew(items, startTime)
		} else {
			result = s.SearchExclusiveAll(items)
		}

		if len(result) > 0 {
			return result
		}

		if timeOut != 0 && time.Now().After(deadline) {
			return nil
		}

		s.cond.Wait()
	}

}

func (s *StreamStore) SearchExclusiveAllNew(items []*Pair[string, string], startTime time.Time) []any {
	result := make([]any, 0)
	for i := range items {
		item := items[i]
		key := item.GetK()
		nested := s.SearchExclusiveNew(key, startTime)
		if len(nested) > 0 {
			result = append(result, []any{key, nested})
		}
	}

	return result

}

func (s *StreamStore) SearchExclusiveAll(items []*Pair[string, string]) []any {

	result := make([]any, 0)
	for i := range items {
		item := items[i]
		key := item.GetK()
		id := item.GetV()

		nested := s.SearchExclusive(key, id)
		if len(nested) > 0 {
			result = append(result, []any{key, nested})
		}
	}

	return result
}

func (s *StreamStore) SearchExclusive(key, targetId string) []any {

	entries := s.store[key]
	n := len(entries)
	l := n
	for i := range len(entries) {
		currentId := entries[i]["id"]
		if compare(currentId, targetId, func(cts, tts int64, cseq, tseq int) bool {
			return (cts > tts) || (cts == tts && cseq > tseq)
		}) {
			l = i
			break
		}
	}

	result := make([]any, 0)

	for l < len(entries) {
		entry := entries[l]
		id := entry["id"]
		other := make([]string, 0)

		for k, v := range entry {
			if k == "id" {
				continue
			}
			other = append(other, k, v)
		}

		data := []any{
			id,
			other,
		}

		result = append(result, data)

		l++
	}
	return result
}

func (s *StreamStore) SearchExclusiveNew(key string, timeStamp time.Time) []any {

	result := make([]any, 0)
	entries := s.store[key]
	for i := range entries {
		entryId := entries[i]["id"]
		inSertionTime := s.timeStore[entryId]

		if inSertionTime.After(timeStamp) {

			data := make([]string, 0)
			for k, v := range entries[i] {
				if k == "id" {
					continue
				}
				data = append(data, k, v)
			}
			result = append(result, []any{entryId, data})
		}
	}

	return result
}

func greaterOrEqualTo(ats, bts int64, seq1, seq2 int) bool {

	// ats -> currentId timeStamp
	// bts -> start timeStamp
	// seq1 -> currentId Seq
	// seq2 -> start Seq

	if ats > bts {
		return true
	}

	return ats == bts && seq1 >= seq2
}

func lesserOrEqualTo(ats, bts int64, seq1, seq2 int) bool {
	// ats -> currentId timeStamp
	// bts -> end timeStamp
	// seq1 -> currentId Seq
	// seq2 -> end Seq

	if ats < bts {
		return true
	}

	return ats == bts && seq1 <= seq2

}

func compare(ats, bts string, fn func(ats, bts int64, seq1, seq2 int) bool) bool {

	ats, aseq, _ := strings.Cut(ats, "-")
	bts, bseq, _ := strings.Cut(bts, "-")
	ts1, _ := strconv.ParseInt(ats, 10, 64) // base 10 and 64 bit integer
	ts2, _ := strconv.ParseInt(bts, 10, 64)
	seq1, _ := strconv.Atoi(aseq)
	seq2, _ := strconv.Atoi(bseq)
	return fn(ts1, ts2, seq1, seq2)
}

package store

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// We are going to use skip list and map
// reason: we want every operation to use approximate O(log(N))

const (
	MaxLevel = 16
	P        = 0.5
)

type SkipNode struct {
	member string
	score  float64
	next   []*SkipNode // forward pointers one per level
	span   []int       // number of L0 nodes we are skipping
}

func NewSkipNode(member string, score float64, level int) *SkipNode {
	return &SkipNode{
		member: member,
		score:  score,
		next:   make([]*SkipNode, level),
		span:   make([]int, level),
	}
}

type SkipList struct {
	head  *SkipNode
	index map[string]*SkipNode
	level int
	size  int
	rand  *rand.Rand // for generating the random level
}

func NewSkipList() *SkipList {
	return &SkipList{
		head:  NewSkipNode("", math.Inf(-1), MaxLevel+1),
		index: make(map[string]*SkipNode),
		level: 0,
		size:  0,
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (sl *SkipList) randomLevel() int {
	lvl := 0
	for sl.rand.Float64() < P && lvl < MaxLevel {
		lvl++
	}
	return lvl
}

func cmp(memberA, memberB string, scoreA, scoreB float64) int {
	if scoreA != scoreB {
		if scoreA < scoreB {
			return -1
		}
		return 1
	}

	return strings.Compare(memberA, memberB)
}

func (sl *SkipList) Insert(member string, score float64) int {

	insertedItems := 1
	if _, ok := sl.index[member]; ok {
		insertedItems--
		sl.Remove(member)
	}

	update := make([]*SkipNode, MaxLevel+1)
	rank := make([]int, MaxLevel+1)
	cur := sl.head

	// 1. Update the latest previous node
	for i := sl.level; i >= 0; i-- {

		if i == sl.level {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for cur.next[i] != nil && cmp(cur.next[i].member, member, cur.next[i].score, score) < 0 {
			rank[i] += cur.span[i]
			cur = cur.next[i]
		}
		update[i] = cur
	}

	// 2. try to create a new level for the new member
	newLevel := sl.randomLevel()
	if newLevel > sl.level {
		for i := sl.level + 1; i <= newLevel; i++ {
			update[i] = sl.head
			sl.head.span[i] = sl.size
			rank[i] = 0
		}
		sl.level = newLevel
	}

	newNode := NewSkipNode(member, score, newLevel+1)

	for i := 0; i <= newLevel; i++ {
		// change the pointer
		previous := update[i]
		newNode.next[i] = previous.next[i]
		previous.next[i] = newNode

		// Update span and rank
		newNode.span[i] = previous.span[i] - (rank[0] - rank[i])
		previous.span[i] = (rank[0] - rank[i]) + 1
	}

	for i := newLevel + 1; i <= sl.level; i++ {
		update[i].span[i]++
	}

	sl.size++
	sl.index[member] = newNode

	return insertedItems
}

func (sl *SkipList) Remove(member string) bool { // O(log(N))

	targetNode, ok := sl.index[member]
	if !ok {
		return false
	}

	update := make([]*SkipNode, MaxLevel+1)
	rank := make([]int, MaxLevel+1)
	cur := sl.head

	for i := sl.level; i >= 0; i-- {
		if i == sl.level {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for cur.next[i] != nil && cmp(cur.next[i].member, targetNode.member, cur.next[i].score, targetNode.score) < 0 {
			rank[i] += cur.span[i]
			cur = cur.next[i]
		}
		update[i] = cur
	}

	node := update[0].next[0]
	for i := 0; i <= sl.level; i++ {
		// cases when the node is present inside the level
		if update[i].next[i] == node {
			update[i].span[i] += node.span[i] - 1 // 1 for removing the target node
			update[i].next[i] = node.next[i]
		} else { // case when node is not present inside the level
			update[i].span[i]--
		}
	}

	// shrink the level
	for sl.level > 0 && sl.head.next[sl.level] == nil {
		sl.level--
	}

	sl.size--
	delete(sl.index, member)
	return true

}

func (sl *SkipList) Rank(member string) int { // O(log(N))

	targetNode, ok := sl.index[member]
	if !ok {
		return -1
	}

	rank := 0
	cur := sl.head

	for i := sl.level; i >= 0; i-- {

		for cur.next[i] != nil && cmp(cur.next[i].member, targetNode.member, cur.next[i].score, targetNode.score) < 0 {
			rank += cur.span[i]
			cur = cur.next[i]
		}
		if cur.next[i] != nil && cmp(cur.next[i].member, targetNode.member, cur.next[i].score, targetNode.score) == 0 {
			rank += cur.span[i]
			break
		}
	}

	return rank - 1

}

func (sl *SkipList) Range(start, end int) []any { // O(log(N) + K )
	if start < 0 {
		start = max(0, sl.size+start)
	}

	if end < 0 {
		end = min(sl.size-1, sl.size+end)
	}

	result := make([]any, 0)

	if start > end {
		return result
	}

	cur := sl.head
	rank := 0

	for i := sl.level; i >= 0; i-- { // O(log(N))
		// why start+1
		// Because rank tracks the head sentinel node which is at position 0 but is not a real node. so real nodes will be always indexed 1
		for cur.next[i] != nil && rank+cur.span[i] <= start+1 {
			rank += cur.span[i]
			cur = cur.next[i]
		}
	}

	for i := start; i <= end && cur != nil; i++ { // O(K)
		result = append(result, cur.member)
		cur = cur.next[0] // at the lower level all the elements will be there
	}

	return result
}

func (sl *SkipList) Score(member string) string {
	node, ok := sl.index[member]
	if !ok {
		return "-1"
	}

	return fmt.Sprintf("%g", node.score)
}

func (sl *SkipList) Size() int { // O(1)
	return sl.size
}

type SkipListBucket struct {
	bucket map[string]*SkipList
}

func NewSkipListBucket() *SkipListBucket {
	return &SkipListBucket{
		bucket: make(map[string]*SkipList),
	}
}

var cSkipListBucket = NewSkipListBucket()

func GetSkipListBucket() *SkipListBucket {
	return cSkipListBucket
}

func (b *SkipListBucket) Insert(set *SetItem) int {

	if b.bucket[set.key] == nil {
		b.bucket[set.key] = NewSkipList()
	}

	skipList := b.bucket[set.key]
	insertedItems := skipList.Insert(set.member, set.score)
	return insertedItems

}

func (b *SkipListBucket) Rank(key, member string) int {
	if b.bucket[key] == nil {
		return -1
	}

	skipList := b.bucket[key]
	return skipList.Rank(member)
}

func (b *SkipListBucket) Range(key string, start, end int) []any {
	if b.bucket[key] == nil {
		return []any{}
	}

	skipList := b.bucket[key]
	return skipList.Range(start, end)
}

func (b *SkipListBucket) Count(key string) int {
	if b.bucket[key] == nil {
		return 0
	}

	skipList := b.bucket[key]
	return skipList.Size()
}

func (b *SkipListBucket) Score(key, member string) string {
	if b.bucket[key] == nil {
		return "-1"
	}

	skipList := b.bucket[key]
	return skipList.Score(member)
}

func (b *SkipListBucket) Exists(key string) bool {
	_, ok := b.bucket[key]
	return ok
}

func (b *SkipListBucket) Remove(key, member string) int {
	if b.bucket[key] == nil {
		return 0
	}

	skipList := b.bucket[key]
	isRemoved := skipList.Remove(member)
	if isRemoved {
		return 1
	}
	return 0
}

type SetItem struct {
	key    string
	member string
	score  float64
}

func NewSetItem(key, score, member string) *SetItem {

	scores, err := strconv.ParseFloat(score, 64)

	if err != nil {
		slog.Error("Unable to convert the score")
		return nil
	}

	slog.Info("Creating a Set Entry with", slog.Any("key", key), slog.Any("score", score), slog.Any("member", member))

	return &SetItem{
		key:    key,
		member: member,
		score:  scores,
	}
}

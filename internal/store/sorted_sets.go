package store

import (
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

	insertedItems := 0
	if _, ok := sl.index[member]; !ok {
		insertedItems++
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

func (sl *SkipList) Rank(member string) int {

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

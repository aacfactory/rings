package rings

import (
	"fmt"
	"github.com/cespare/xxhash/v2"
	"sort"
	"sync"
)

const (
	maxEntries = uint64(256)
)

type HashRingEntry[E Entry] struct {
	Entry  E
	Active bool
	Low    uint64
	High   uint64
}

func (entry *HashRingEntry[E]) String() (s string) {
	active := "F"
	if entry.Active {
		active = "T"
	}
	s = fmt.Sprintf("{%s:[%d, %d) (%s)}", entry.Entry.Key(), entry.Low, entry.High, active)
	return
}

func (entry *HashRingEntry[E]) Less(o *HashRingEntry[E]) bool {
	if entry.High < o.High {
		return true
	}
	return entry.High == o.High && entry.Low < o.Low
}

func (entry *HashRingEntry[E]) RangeSize() uint64 {
	return entry.High - entry.Low
}

type RangeSizeSortedHashRingEntries[E Entry] []*HashRingEntry[E]

func (entries RangeSizeSortedHashRingEntries[E]) Len() int {
	return len(entries)
}

func (entries RangeSizeSortedHashRingEntries[E]) Less(i, j int) bool {
	return entries[i].RangeSize() < entries[j].RangeSize() && entries[j].Active
}

func (entries RangeSizeSortedHashRingEntries[E]) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
	return
}

type HashRingEntries[E Entry] []*HashRingEntry[E]

func (entries HashRingEntries[E]) Len() int {
	return len(entries)
}

func (entries HashRingEntries[E]) Less(i, j int) bool {
	return entries[i].Less(entries[j])
}

func (entries HashRingEntries[E]) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
	return
}

func (entries HashRingEntries[E]) BiggestRangeAndActiveOne() (entry *HashRingEntry[E]) {
	sorted := RangeSizeSortedHashRingEntries[E](entries)
	sort.Sort(sorted)
	target := sorted[sorted.Len()-1]
	sort.Sort(entries)
	if !target.Active {
		return
	}
	entry = target
	return
}

func (entries HashRingEntries[E]) Get(n uint64) (entry *HashRingEntry[E], has bool) {
	idx := -1
	for i, hashed := range entries {
		if hashed.Low <= n && hashed.High > n {
			entry = hashed
			idx = i
			break
		}
	}
	entry = entries[idx]
	if entry.Active {
		has = true
		return
	}
	for {
		idx--
		if idx < 0 {
			break
		}
		entry = entries[idx]
		if entry.Active {
			has = true
			return
		}
	}
	return
}

func NewHashed[E Entry](entries ...E) (v *HashRing[E]) {
	hashedEntries := make([]*HashRingEntry[E], 0, 1)
	if entries != nil && len(entries) > 0 {
		for _, entry := range entries {
			hashedEntries = append(hashedEntries, &HashRingEntry[E]{
				Entry:  entry,
				Active: true,
				Low:    0,
				High:   0,
			})
		}
		span := maxEntries / uint64(len(hashedEntries))
		for i, entry := range hashedEntries {
			entry.Low = span * uint64(i)
			entry.High = span * uint64(i+1)
		}
		hashedEntries[len(entries)-1].High = maxEntries
	}
	v = &HashRing[E]{
		locker:  sync.RWMutex{},
		entries: hashedEntries,
	}
	return
}

type HashRing[E Entry] struct {
	locker  sync.RWMutex
	entries HashRingEntries[E]
}

func (r *HashRing[E]) Get(key []byte) (entry E, has bool) {
	r.locker.RLock()
	idx := xxhash.Sum64(key) % maxEntries
	hashed, hasHashed := r.entries.Get(idx)
	if !hasHashed {
		r.locker.RUnlock()
		return
	}
	entry = hashed.Entry
	has = true
	r.locker.RUnlock()
	return
}

func (r *HashRing[E]) Add(entry E) (prevActive E, cLow uint64, cHigh uint64, active func(), cancel func(), ok bool) {
	r.locker.Lock()
	defer r.locker.Unlock()
	if uint64(r.entries.Len()) >= maxEntries {
		return
	}
	var prevActiveHashed *HashRingEntry[E]
	hashed := &HashRingEntry[E]{
		Entry:  entry,
		Active: false,
		Low:    0,
		High:   maxEntries,
	}
	if r.entries.Len() == 0 {
		r.entries = append(r.entries, hashed)
	} else {
		prevActiveHashed = r.entries.BiggestRangeAndActiveOne()
		if prevActiveHashed == nil {
			return
		}
		hashed.Low = (prevActiveHashed.High-prevActiveHashed.Low)/2 + prevActiveHashed.Low
		hashed.High = prevActiveHashed.High
		prevActiveHashed.High = hashed.Low
		prevActive = prevActiveHashed.Entry
		r.entries = append(r.entries, hashed)
		sort.Sort(r.entries)
	}
	cLow = hashed.Low
	cHigh = hashed.High
	active = func() {
		r.locker.Lock()
		hashed.Active = true
		r.locker.Unlock()
	}
	cancel = func() {
		r.locker.Lock()
		if prevActiveHashed != nil {
			prevActiveHashed.High = hashed.High
		}
		entries := make([]*HashRingEntry[E], 0, r.entries.Len()-1)
		for _, hashedEntry := range r.entries {
			if hashedEntry.Entry.Key() == entry.Key() {
				continue
			}
			entries = append(entries, hashedEntry)
		}
		r.entries = entries
		sort.Sort(r.entries)
		r.locker.Unlock()
	}
	ok = true
	return
}

func (r *HashRing[E]) AddDeclared(entry E, low uint64, high uint64) (ok bool) {
	r.locker.Lock()
	defer r.locker.Unlock()
	if uint64(r.entries.Len()) >= maxEntries {
		return
	}
	hashed := &HashRingEntry[E]{
		Entry:  entry,
		Active: true,
		Low:    low,
		High:   high,
	}
	if r.entries.Len() == 0 {
		r.entries = append(r.entries, hashed)
	} else {
		for _, e := range r.entries {
			if intersect(e.Low, e.High, low, high) {
				return
			}
		}
		r.entries = append(r.entries, hashed)
		sort.Sort(r.entries)
	}
	ok = true
	return
}

func (r *HashRing[E]) Size() (n int) {
	n = len(r.entries)
	return
}

func (r *HashRing[E]) State(key string) (active bool, low uint64, high uint64, has bool) {
	for _, entry := range r.entries {
		if entry.Entry.Key() == key {
			active = entry.Active
			low = entry.Low
			high = entry.High
			has = true
			return
		}
	}
	return
}

func (r *HashRing[E]) States(fn func(key string, active bool, low uint64, high uint64) bool) {
	for _, entry := range r.entries {
		if !fn(entry.Entry.Key(), entry.Active, entry.Low, entry.High) {
			break
		}
	}
	return
}

func (r *HashRing[E]) String() (s string) {
	for _, entry := range r.entries {
		s = s + ", " + entry.String()
	}
	if s != "" {
		s = s[2:]
	}
	s = "[" + s + "]"
	return
}

func intersect(sLow uint64, sHigh uint64, tLow uint64, tHigh uint64) (ok bool) {
	for s := sLow; s < sHigh; s++ {
		for t := tLow; t < tHigh; t++ {
			if s == t {
				ok = true
				return
			}
		}
	}
	return
}

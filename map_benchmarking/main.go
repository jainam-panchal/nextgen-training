package main

import (
	"math/bits"

	"github.com/RoaringBitmap/roaring/v2"
)

type IntSet interface {
	Add(x uint32)
	Contains(x uint32) bool
	Remove(x uint32)
	Len() uint64
	Iterate(yield func(uint32) bool)
	MemoryBytes() uint64
}

type GoSet struct {
	m map[uint32]struct{}
}

func NewGoSet() *GoSet {
	return &GoSet{m: make(map[uint32]struct{})}
}

func (s *GoSet) Add(x uint32) {
	s.m[x] = struct{}{}
}

func (s *GoSet) Contains(x uint32) bool {
	_, ok := s.m[x]
	return ok
}

func (s *GoSet) Remove(x uint32) {
	delete(s.m, x)
}

func (s *GoSet) Len() uint64 {
	return uint64(len(s.m))
}

func (s *GoSet) Iterate(yield func(uint32) bool) {
	for k := range s.m {
		if !yield(k) {
			return
		}
	}
}

func (s *GoSet) MemoryBytes() uint64 {
	return 0
}

type RoaringSet struct {
	b *roaring.Bitmap
}

func NewRoaringSet() *RoaringSet {
	return &RoaringSet{b: roaring.New()}
}

func (s *RoaringSet) Add(x uint32) {
	s.b.Add(x)
}

func (s *RoaringSet) Contains(x uint32) bool {
	return s.b.Contains(x)
}

func (s *RoaringSet) Remove(x uint32) {
	s.b.Remove(x)
}

func (s *RoaringSet) Len() uint64 {
	return s.b.GetCardinality()
}

func (s *RoaringSet) Iterate(yield func(uint32) bool) {
	it := s.b.Iterator()
	for it.HasNext() {
		if !yield(it.Next()) {
			return
		}
	}
}

func (s *RoaringSet) MemoryBytes() uint64 {
	return s.b.GetSizeInBytes()
}

type BitArraySet struct {
	bits  []uint64
	count int
}

func NewBitArraySet() *BitArraySet {
	return &BitArraySet{bits: make([]uint64, 16)}
}

func (s *BitArraySet) Add(x uint32) {
	idx := int(x / 64)
	bit := uint64(1) << (x % 64)
	if idx >= len(s.bits) {
		newBits := make([]uint64, idx+1)
		copy(newBits, s.bits)
		s.bits = newBits
	}
	if s.bits[idx]&bit == 0 {
		s.bits[idx] |= bit
		s.count++
	}
}

func (s *BitArraySet) Contains(x uint32) bool {
	idx := int(x / 64)
	if idx >= len(s.bits) {
		return false
	}
	return s.bits[idx]&(uint64(1)<<(x%64)) != 0
}

func (s *BitArraySet) Remove(x uint32) {
	idx := int(x / 64)
	if idx >= len(s.bits) {
		return
	}
	bit := uint64(1) << (x % 64)
	if s.bits[idx]&bit != 0 {
		s.bits[idx] &^= bit
		s.count--
	}
}

func (s *BitArraySet) Len() uint64 {
	return uint64(s.count)
}

func (s *BitArraySet) Iterate(yield func(uint32) bool) {
	for i, w := range s.bits {
		if w == 0 {
			continue
		}
		base := uint32(i) * 64
		for j := uint32(0); j < 64; j++ {
			if w&(1<<j) != 0 {
				if !yield(base + j) {
					return
				}
			}
		}
	}
}

func (s *BitArraySet) MemoryBytes() uint64 {
	return uint64(cap(s.bits)) * 8
}

func bitArrayUnion(a, b *BitArraySet) *BitArraySet {
	maxLen := len(a.bits)
	if len(b.bits) > maxLen {
		maxLen = len(b.bits)
	}
	u := make([]uint64, maxLen)
	count := 0
	for i := 0; i < maxLen; i++ {
		var va, vb uint64
		if i < len(a.bits) {
			va = a.bits[i]
		}
		if i < len(b.bits) {
			vb = b.bits[i]
		}
		u[i] = va | vb
		count += bits.OnesCount64(u[i])
	}
	return &BitArraySet{bits: u, count: count}
}

func bitArrayIntersection(a, b *BitArraySet) *BitArraySet {
	maxLen := len(a.bits)
	if len(b.bits) < maxLen {
		maxLen = len(b.bits)
	}
	u := make([]uint64, maxLen)
	count := 0
	for i := 0; i < maxLen; i++ {
		u[i] = a.bits[i] & b.bits[i]
		count += bits.OnesCount64(u[i])
	}
	return &BitArraySet{bits: u, count: count}
}

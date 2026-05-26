package main

import (
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

package main

import (
	"fmt"
	"testing"
)

// -------- linked list --------

type Node struct {
	val  int
	next *Node
}

type List struct {
	head *Node
	tail *Node
	len  int
}

func newList() *List {
	return &List{}
}

func (l *List) append(val int) {
	n := &Node{val: val}
	if l.tail == nil {
		l.head = n
		l.tail = n
	} else {
		l.tail.next = n
		l.tail = n
	}
	l.len++
}

func (l *List) prepend(val int) {
	l.head = &Node{val: val, next: l.head}
	if l.tail == nil {
		l.tail = l.head
	}
	l.len++
}

func (l *List) insertFront(val int) {
	l.prepend(val)
}

func (l *List) deleteFront() {
	if l.head == nil {
		return
	}
	l.head = l.head.next
	if l.head == nil {
		l.tail = nil
	}
	l.len--
}

func (l *List) get(idx int) int {
	p := l.head
	for i := 0; i < idx; i++ {
		p = p.next
	}
	return p.val
}

func (l *List) iterate(fn func(int)) {
	for p := l.head; p != nil; p = p.next {
		fn(p.val)
	}
}

// -------- sink --------

var sink int

// -------- test sizes --------

var ns = []int{10, 100, 1_000, 10_000, 100_000}

// -------- benchmark: build by appending --------

func BenchmarkAppend(b *testing.B) {
	for _, n := range ns {
		vals := make([]int, n)
		for i := range vals {
			vals[i] = i
		}

		b.Run(fmt.Sprintf("Slice/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s := make([]int, 0, n)
				for _, v := range vals {
					s = append(s, v)
				}
				sink += len(s)
			}
		})

		b.Run(fmt.Sprintf("LinkedList/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				l := newList()
				for _, v := range vals {
					l.append(v)
				}
				sink += l.len
			}
		})
	}
}

// -------- benchmark: sequential iteration --------

func BenchmarkIterate(b *testing.B) {
	for _, n := range ns {
		vals := make([]int, n)
		for i := range vals {
			vals[i] = i
		}

		l := newList()
		for _, v := range vals {
			l.append(v)
		}

		b.Run(fmt.Sprintf("Slice/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			var local int
			for i := 0; i < b.N; i++ {
				for _, v := range vals {
					local += v
				}
			}
			sink += local
		})

		b.Run(fmt.Sprintf("LinkedList/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			var local int
			for i := 0; i < b.N; i++ {
				l.iterate(func(v int) {
					local += v
				})
			}
			sink += local
		})
	}
}

// -------- benchmark: random access --------

func BenchmarkRandomAccess(b *testing.B) {
	for _, n := range ns {
		vals := make([]int, n)
		for i := range vals {
			vals[i] = i
		}

		l := newList()
		for _, v := range vals {
			l.append(v)
		}

		indices := make([]int, n)
		for i := range indices {
			indices[i] = n - 1 - i
		}

		b.Run(fmt.Sprintf("Slice/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			var local int
			for i := 0; i < b.N; i++ {
				for _, idx := range indices {
					local += vals[idx]
				}
			}
			sink += local
		})

		b.Run(fmt.Sprintf("LinkedList/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			var local int
			for i := 0; i < b.N; i++ {
				for _, idx := range indices {
					local += l.get(idx)
				}
			}
			sink += local
		})
	}
}

// -------- benchmark: insert at front --------

func BenchmarkInsertFront(b *testing.B) {
	for _, n := range ns {
		vals := make([]int, n)
		for i := range vals {
			vals[i] = i
		}

		b.Run(fmt.Sprintf("Slice/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s := make([]int, 0, n)
				for _, v := range vals {
					s = append([]int{v}, s...)
				}
				sink += len(s)
			}
		})

		b.Run(fmt.Sprintf("LinkedList/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				l := newList()
				for _, v := range vals {
					l.insertFront(v)
				}
				sink += l.len
			}
		})
	}
}

// -------- benchmark: delete from front --------

func BenchmarkDeleteFront(b *testing.B) {
	for _, n := range ns {
		vals := make([]int, n)
		for i := range vals {
			vals[i] = i
		}

		b.Run(fmt.Sprintf("Slice/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s := make([]int, n)
				copy(s, vals)
				for len(s) > 0 {
					s = s[1:]
				}
				sink += len(s)
			}
		})

		b.Run(fmt.Sprintf("LinkedList/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				l := newList()
				for _, v := range vals {
					l.append(v)
				}
				for l.head != nil {
					l.deleteFront()
				}
				sink += l.len
			}
		})
	}
}

// -------- benchmark: build by prepending --------

func BenchmarkPrependBuild(b *testing.B) {
	for _, n := range ns {
		vals := make([]int, n)
		for i := range vals {
			vals[i] = i
		}

		b.Run(fmt.Sprintf("Slice/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var s []int
				for _, v := range vals {
					s = append([]int{v}, s...)
				}
				sink += len(s)
			}
		})

		b.Run(fmt.Sprintf("LinkedList/n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				l := newList()
				for _, v := range vals {
					l.prepend(v)
				}
				sink += l.len
			}
		})
	}
}

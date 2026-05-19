package btree

import "testing"

func BenchmarkBTreeInsert(b *testing.B) {
	for _, size := range []int{1_000, 10_000, 100_000} {
		b.Run("n="+itoa(size), func(b *testing.B) {
			keys := makeSequential(size)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tree := New[int, int](32)
				for _, k := range keys {
					tree.Insert(k, k)
				}
			}
		})
	}
}

func BenchmarkBTreeSearch(b *testing.B) {
	for _, size := range []int{1_000, 10_000, 100_000} {
		b.Run("n="+itoa(size), func(b *testing.B) {
			tree := New[int, int](32)
			keys := makeSequential(size)
			for _, k := range keys {
				tree.Insert(k, k)
			}

			idx := 0
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = tree.Search(keys[idx])
				idx++
				if idx == len(keys) {
					idx = 0
				}
			}
		})
	}
}

func BenchmarkBTreeRangeQuery(b *testing.B) {
	for _, size := range []int{1_000, 10_000, 100_000} {
		b.Run("n="+itoa(size), func(b *testing.B) {
			tree := New[int, int](32)
			keys := makeSequential(size)
			for _, k := range keys {
				tree.Insert(k, k)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				min := (i * 97) % size
				max := min + 250
				if max >= size {
					max = size - 1
				}
				_ = tree.RangeQuery(min, max)
			}
		})
	}
}

func BenchmarkBTreeDelete(b *testing.B) {
	for _, size := range []int{1_000, 10_000, 100_000} {
		b.Run("n="+itoa(size), func(b *testing.B) {
			keys := makeSequential(size)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tree := New[int, int](32)
				for _, k := range keys {
					tree.Insert(k, k)
				}

				target := keys[(i*37)%len(keys)]
				_ = tree.Delete(target)
			}
		})
	}
}

func makeSequential(size int) []int {
	keys := make([]int, size)
	for i := range keys {
		keys[i] = i
	}
	return keys
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := [20]byte{}
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}

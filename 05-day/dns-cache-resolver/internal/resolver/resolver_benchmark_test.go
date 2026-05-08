package resolver

import (
	cachestore "dns-cache-resolver/internal/cache_store"
	dnsrecord "dns-cache-resolver/internal/dns_record"
	"dns-cache-resolver/internal/upstream"
	"fmt"
	"testing"
)

func BenchmarkResolve(b *testing.B) {
	sizes := []int{100, 1_000, 10_000, 100_000, 500_000, 1_000_000}

	for _, n := range sizes {
		domains := make([]string, n)
		for i := 0; i < n; i++ {
			domains[i] = fmt.Sprintf("host-%d.example.com", i)
		}

		b.Run(fmt.Sprintf("CustomMapStore/n=%d", n), func(b *testing.B) {
			store := cachestore.NewCustomMapStore[dnsrecord.DNSRecord]()
			r := NewResolver(store, upstream.NewDummyUpstreamResolver())

			// Warm-up fills cache so benchmark focuses on resolve path under size n.
			for _, d := range domains {
				_, _ = r.Resolve(d)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				domain := domains[i%n]
				_, _ = r.Resolve(domain)
			}
		})

		b.Run(fmt.Sprintf("InbuiltMapStore/n=%d", n), func(b *testing.B) {
			store := cachestore.NewInbuiltMapStore[dnsrecord.DNSRecord]()
			r := NewResolver(store, upstream.NewDummyUpstreamResolver())

			// Warm-up fills cache so benchmark focuses on resolve path under size n.
			for _, d := range domains {
				_, _ = r.Resolve(d)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				domain := domains[i%n]
				_, _ = r.Resolve(domain)
			}
		})
	}
}

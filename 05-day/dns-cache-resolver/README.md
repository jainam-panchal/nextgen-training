## problem statement

Requirements:
Build a DNS cache resolver in Go:

1. STRUCTURE:

- Implement your OWN hash map from scratch first (don't use Go's built-
  in map)
- Custom hash function for strings (e.g., FNV-1a or djb2)
- Bucket array with chaining (linked list per bucket)
- Dynamic resizing when load factor > 0.75
- DNSRecord: { Domain, IP, TTL time.Duration, CreatedAt time.Time,
  HitCount int }

2. CORE FEATURES:- Resolve(domain) → IP: check cache first, simulate upstream lookup if
   miss

- AddRecord(domain, ip, ttl): insert with expiration
- Eviction: remove expired entries (TTL-based)
- Cache stats: hit rate, miss rate, total entries, memory estimate
- Wildcard matching: \*.example.com should match sub.example.com

3. AFTER CUSTOM IMPLEMENTATION:

- Re-implement using Go's built-in map[string]\*DNSRecord
- Benchmark both: go test -bench=BenchmarkResolve -benchmem
- Compare: ops/sec, memory per entry, resize behavior

4. GO REQUIREMENTS:

- Implement the hash function — explain why you chose it
- Handle hash collisions — demonstrate with intentionally colliding
  keys
- Use time.Ticker for periodic TTL cleanup goroutine (preview
  concurrency)
- JSON struct tags for serialization

5. TTL IMPLEMENTATION:

- Each record expires after its TTL
- Background cleanup every 30 seconds removes expired entries
- Lazy expiration: also check on lookup

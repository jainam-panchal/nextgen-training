before map resizing 
```bash
(pprof) top                                                                                                
Showing nodes accounting for 4420ms, 73.42% of 6020ms total
Dropped 67 nodes (cum <= 30.10ms)
Showing top 10 nodes out of 59
      flat  flat%   sum%        cum   cum%
    1860ms 30.90% 30.90%     3770ms 62.62%  runtime.mapassign_faststr
     600ms  9.97% 40.86%      600ms  9.97%  internal/runtime/maps.ctrlGroup.matchH2 (inline)
     420ms  6.98% 47.84%      420ms  6.98%  memeqbody
     390ms  6.48% 54.32%      390ms  6.48%  aeshashbody
     250ms  4.15% 58.47%      570ms  9.47%  bufio.ScanWords
     250ms  4.15% 62.62%      550ms  9.14%  runtime.mallocgcTiny
     180ms  2.99% 65.61%      270ms  4.49%  runtime.tryDeferToSpanScan
     170ms  2.82% 68.44%      180ms  2.99%  bufio.isSpace (inline)
     160ms  2.66% 71.10%      760ms 12.62%  runtime.mallocgc
     140ms  2.33% 73.42%      970ms 16.11%  bufio.(*Scanner).Scan
(pprof) exit
jainam-panchal@mindarray:~/repos/nextgen-training/10-day/snippets/code_3$ go tool pprof mem.prof
File: main
Build ID: 771dd239c87c6a7f81aba7fd19dd0a6aee9caeb0
Type: inuse_space
Time: 2026-05-17 13:58:19 IST
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 5800.28kB, 100% of 5800.28kB total
Showing top 10 nodes out of 23
      flat  flat%   sum%        cum   cum%
    4104kB 70.76% 70.76%     4104kB 70.76%  runtime.mallocgc
 1184.27kB 20.42% 91.17%  1184.27kB 20.42%  runtime/pprof.StartCPUProfile
     512kB  8.83%   100%      512kB  8.83%  os.newFile
         0     0%   100%  1696.28kB 29.24%  main.main
         0     0%   100%      512kB  8.83%  os.Create (inline)
         0     0%   100%      512kB  8.83%  os.OpenFile
         0     0%   100%      512kB  8.83%  os.openFileNolog
         0     0%   100%     4104kB 70.76%  runtime.allocm
         0     0%   100%      513kB  8.84%  runtime.handoffp
         0     0%   100%  1696.28kB 29.24%  runtime.main
(pprof) 
```


after precomputed map
```bash
jainam-panchal@mindarray:~/repos/nextgen-training/10-day/snippets/code_3$ go tool pprof cpu_opti.prof 
File: main
Build ID: bcab78dc82bba07ed39f2255c05cbec7d5798a2f
Type: cpu
Time: 2026-05-17 14:03:36 IST
Duration: 431.95ms, Total samples = 4.32s (1000.11%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 3110ms, 71.99% of 4320ms total
Dropped 40 nodes (cum <= 21.60ms)
Showing top 10 nodes out of 60
      flat  flat%   sum%        cum   cum%
     810ms 18.75% 18.75%     1990ms 46.06%  runtime.mapassign_faststr
     440ms 10.19% 28.94%      440ms 10.19%  internal/runtime/maps.ctrlGroup.matchH2 (inline)
     320ms  7.41% 36.34%      320ms  7.41%  memeqbody
     300ms  6.94% 43.29%      720ms 16.67%  bufio.ScanWords
     270ms  6.25% 49.54%      270ms  6.25%  bufio.isSpace (inline)
     270ms  6.25% 55.79%      650ms 15.05%  runtime.mallocgcTiny
     250ms  5.79% 61.57%      250ms  5.79%  aeshashbody
     160ms  3.70% 65.28%      160ms  3.70%  runtime.nextFreeFast (inline)
     150ms  3.47% 68.75%      150ms  3.47%  unicode/utf8.DecodeRune (inline)
     140ms  3.24% 71.99%      140ms  3.24%  bufio.NewScanner
(pprof) jainam-panchal@mindarray:~/repos/nextgen-training/10-day/snippets/code_3$ go tool pprof mem_opti.prof 
File: main
Build ID: bcab78dc82bba07ed39f2255c05cbec7d5798a2f
Type: inuse_space
Time: 2026-05-17 14:03:36 IST
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 4774.35kB, 100% of 4774.35kB total
Showing top 10 nodes out of 19
      flat  flat%   sum%        cum   cum%
    3078kB 64.47% 64.47%     3078kB 64.47%  runtime.mallocgc
 1184.27kB 24.80% 89.27%  1184.27kB 24.80%  runtime/pprof.StartCPUProfile
  512.08kB 10.73%   100%   512.08kB 10.73%  compress/gzip.NewWriterLevel
         0     0%   100%  1184.27kB 24.80%  main.main
         0     0%   100%     3078kB 64.47%  runtime.allocm
         0     0%   100%  1184.27kB 24.80%  runtime.main
         0     0%   100%     1026kB 21.49%  runtime.mcall
         0     0%   100%     2052kB 42.98%  runtime.mstart
         0     0%   100%     2052kB 42.98%  runtime.mstart0
         0     0%   100%     2052kB 42.98%  runtime.mstart1
(pprof) 
```
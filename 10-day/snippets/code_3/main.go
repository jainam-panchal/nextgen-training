package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

func main() {
	flag.Parse()

	files := flag.Args()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}

		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}

		defer f.Close()
		defer pprof.StopCPUProfile()
	}

	if len(files) == 0 {
		log.Fatal("no files to process")
	}

	start := time.Now()
	workerCount := runtime.NumCPU()

	// channels
	fileCh := make(chan string)
	resultCh := make(chan map[string]int)

	var wg sync.WaitGroup

	// final merged result
	finalResult := make(map[string]int, 100_000)

	// 1. Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(fileCh, resultCh, &wg)
	}

	// 2. Send files into fileCh
	go func() {
		for _, file := range files {
			fileCh <- file
		}
		close(fileCh)
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 4. Merge worker results
	for result := range resultCh {
		for k, v := range result {
			finalResult[k] += v
		}
	}

	fmt.Printf("Processing took: %v\n", time.Since(start))

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		runtime.GC()

		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
	}
}

func worker(fileCh <-chan string, resultCh chan<- map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()

	localResult := make(map[string]int, 100_000)

	for fileName := range fileCh {
		processFile(localResult, fileName)
	}
	resultCh <- localResult
}

func processFile(localResult map[string]int, fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		localResult[word]++
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

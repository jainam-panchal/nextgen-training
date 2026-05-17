package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"
)

var (
	workersFlag = flag.String("workers", "", "provide worker count")
	cpuproFile  = flag.String("cpuprofile", "", "cpuprof file")
	memprofFile = flag.String("memprofile", "", "memprof file")
)

func worker(wg *sync.WaitGroup, fileChan chan string, resultChan chan map[string]int) {
	defer wg.Done()

	localMap := make(map[string]int)

	for file := range fileChan {
		process(file, localMap)
	}

	resultChan <- localMap
}

func process(fileName string, mp map[string]int) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		word := scanner.Text()
		mp[word]++
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

func main() {
	flag.Parse()

	// cpu profile
	if *cpuproFile != "" {
		f, err := os.Create(*cpuproFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

	if *workersFlag == "" {
		log.Fatal("usage: program -workers <count> <file1> [file2...]")
	}

	cores := runtime.NumCPU()
	maxWorkers, err := strconv.Atoi(*workersFlag)

	if err != nil || maxWorkers < 1 {
		log.Fatal("workers must be a positive integer")
	}
	if cores < maxWorkers {
		maxWorkers = cores
	}

	files := flag.Args()
	if len(files) < 1 {
		log.Fatal("provide at least one file")
	}

	var wg sync.WaitGroup
	fileChan := make(chan string)
	resultChan := make(chan map[string]int)

	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	go func() {
		for _, file := range files {
			fileChan <- file
		}
		close(fileChan)
	}()

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(&wg, fileChan, resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	mp := make(map[string]int)
	for resultMap := range resultChan {
		for key, val := range resultMap {
			mp[key] += val
		}
	}

	if *memprofFile != "" {
		f, err := os.Create(*memprofFile)
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

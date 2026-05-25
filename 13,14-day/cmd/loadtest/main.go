package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var (
		baseURL        = flag.String("base-url", "http://localhost:8081", "target API base URL")
		userCount      = flag.Int("users", 50, "number of bidder users")
		itemCount      = flag.Int("items", 5, "number of auction items")
		bidsPerUser    = flag.Int("bids-per-user", 10, "number of bids each user attempts")
		duration       = flag.Duration("duration", 30*time.Second, "test duration")
		withAuthHeader = flag.Bool("with-auth-header", true, "send X-User-ID header (enables rate limiter)")
		throughputMode = flag.Bool("throughput-mode", false, "disable obvious business-logic failures (no auth header + one item per user)")
	)
	flag.Parse()
	if *throughputMode {
		*withAuthHeader = false
		if *itemCount < *userCount {
			*itemCount = *userCount
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}

	sellerID := createUser(client, *baseURL, "seller", 1_000_000)
	bidderIDs := make([]int, 0, *userCount)
	for i := 0; i < *userCount; i++ {
		bidderIDs = append(bidderIDs, createUser(client, *baseURL, fmt.Sprintf("bidder-%d", i), 1_000_000))
	}

	itemIDs := make([]int, 0, *itemCount)
	now := time.Now().UTC()
	for i := 0; i < *itemCount; i++ {
		itemIDs = append(itemIDs, createItem(
			client, *baseURL,
			fmt.Sprintf("item-%d", i),
			[]string{"Electronics", "Phones"},
			"load test item",
			sellerID,
			100,
			now.Add(-time.Minute),
			now.Add(2*time.Hour),
		))
	}

	seq := make([]atomic.Int64, *itemCount)
	for i := 0; i < *itemCount; i++ {
		seq[i].Store(100)
	}

	interval := *duration / time.Duration(*bidsPerUser)
	var wg sync.WaitGroup
	var okCount atomic.Int64
	var failCount atomic.Int64
	start := time.Now()

	wg.Add(*userCount)
	for i := 0; i < *userCount; i++ {
		userID := bidderIDs[i]
		userIdx := i
		go func(uid int) {
			defer wg.Done()
			for k := 0; k < *bidsPerUser; k++ {
				itemIdx := (uid + k) % *itemCount
				if *throughputMode {
					// Isolate each user to a dedicated item to remove bid-race conflicts.
					itemIdx = userIdx % *itemCount
				}
				itemID := itemIDs[itemIdx]
				amount := float64(seq[itemIdx].Add(1))
				if placeBid(client, *baseURL, itemID, uid, amount, *withAuthHeader) {
					okCount.Add(1)
				} else {
					failCount.Add(1)
				}
				time.Sleep(interval)
			}
		}(userID)
	}
	wg.Wait()
	elapsed := time.Since(start)

	for _, itemID := range itemIDs {
		endAuction(client, *baseURL, itemID)
	}

	total := okCount.Load()
	failed := failCount.Load()
	attempted := int64(*userCount * *bidsPerUser)

	log.Printf(
		"loadtest complete: users=%d items=%d bids/user=%d attempted=%d successful=%d failed=%d elapsed=%s throughput=%.2f ok_bids/sec with_auth_header=%t throughput_mode=%t",
		*userCount, *itemCount, *bidsPerUser, attempted, total, failed, elapsed, float64(total)/elapsed.Seconds(), *withAuthHeader, *throughputMode,
	)
}

func createUser(client *http.Client, baseURL, name string, balance float64) int {
	payload := map[string]any{"name": name, "balance": balance}
	var out map[string]any
	postJSON(client, baseURL, "/users", payload, &out)
	return int(out["user_id"].(float64))
}

func createItem(client *http.Client, baseURL, name string, path []string, description string, sellerID int, startPrice float64, startTime, endTime time.Time) int {
	payload := map[string]any{
		"name":          name,
		"category_path": path,
		"description":   description,
		"seller_id":     sellerID,
		"start_price":   startPrice,
		"start_time":    startTime.Format(time.RFC3339),
		"end_time":      endTime.Format(time.RFC3339),
	}
	var out map[string]any
	postJSON(client, baseURL, "/items", payload, &out)
	return int(out["item_id"].(float64))
}

func placeBid(client *http.Client, baseURL string, itemID, userID int, amount float64, withAuthHeader bool) bool {
	payload := map[string]any{"user_id": userID, "amount": amount}
	reqBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, baseURL+fmt.Sprintf("/items/%d/bid", itemID), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if withAuthHeader {
		req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusCreated
}

func endAuction(client *http.Client, baseURL string, itemID int) {
	req, _ := http.NewRequest(http.MethodPost, baseURL+fmt.Sprintf("/items/%d/end", itemID), nil)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func postJSON(client *http.Client, baseURL, path string, payload any, out any) {
	body, _ := json.Marshal(payload)
	resp, err := client.Post(baseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("post %s failed: %v", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		log.Fatalf("post %s status=%d body=%s", path, resp.StatusCode, string(b))
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			log.Fatalf("decode %s failed: %v", path, err)
		}
	}
}

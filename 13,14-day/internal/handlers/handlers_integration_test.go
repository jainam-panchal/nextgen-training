package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"realtime-auction/internal/auction"
)

func TestAuctionEndpointsIntegration(t *testing.T) {
	store := auction.NewAuctionStore()
	svc := auction.NewAuctionService(store)
	h := NewAuctionHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	userID := postJSONAndGetID(t, srv.URL+"/users", map[string]any{
		"name":    "seller",
		"balance": 10000,
	}, "user_id")

	start := time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)
	end := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	itemID := postJSONAndGetID(t, srv.URL+"/items", map[string]any{
		"name":          "iPhone",
		"category_path": []string{"Electronics", "Phones"},
		"description":   "demo",
		"seller_id":     userID,
		"start_price":   100,
		"start_time":    start,
		"end_time":      end,
	}, "item_id")

	bidderID := postJSONAndGetID(t, srv.URL+"/users", map[string]any{
		"name":    "bidder",
		"balance": 10000,
	}, "user_id")

	postJSONAndGetID(t, srv.URL+"/items/"+itoa(itemID)+"/bid", map[string]any{
		"user_id": bidderID,
		"amount":  200,
	}, "bid_id")

	resp, err := http.Get(srv.URL + "/items/" + itoa(itemID))
	if err != nil {
		t.Fatalf("get item failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/items?category=Electronics")
	if err != nil {
		t.Fatalf("browse category failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestSSELiveUpdates(t *testing.T) {
	store := auction.NewAuctionStore()
	svc := auction.NewAuctionService(store)
	h := NewAuctionHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sellerID := postJSONAndGetID(t, srv.URL+"/users", map[string]any{
		"name":    "seller",
		"balance": 10000,
	}, "user_id")
	bidderID := postJSONAndGetID(t, srv.URL+"/users", map[string]any{
		"name":    "bidder",
		"balance": 10000,
	}, "user_id")
	start := time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)
	end := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	itemID := postJSONAndGetID(t, srv.URL+"/items", map[string]any{
		"name":          "MacBook",
		"category_path": []string{"Electronics", "Laptops"},
		"description":   "demo",
		"seller_id":     sellerID,
		"start_price":   100,
		"start_time":    start,
		"end_time":      end,
	}, "item_id")

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/items/"+itoa(itemID)+"/live", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("sse connect failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	done := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "data: ") {
				done <- line
				return
			}
		}
		done <- ""
	}()

	postJSONAndGetID(t, srv.URL+"/items/"+itoa(itemID)+"/bid", map[string]any{
		"user_id": bidderID,
		"amount":  250,
	}, "bid_id")

	select {
	case line := <-done:
		if line == "" {
			t.Fatal("expected SSE data line")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for SSE event")
	}
}

func postJSONAndGetID(t *testing.T, url string, payload map[string]any, field string) int {
	t.Helper()
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d body=%s", resp.StatusCode, string(b))
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	v, ok := out[field]
	if !ok {
		t.Fatalf("missing field %s", field)
	}
	return int(v.(float64))
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

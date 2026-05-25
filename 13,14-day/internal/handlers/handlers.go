package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"realtime-auction/internal/auction"
	"realtime-auction/internal/models"
)

type AuctionHandler struct {
	service *auction.AuctionService
}

func NewAuctionHandler(service *auction.AuctionService) *AuctionHandler {
	return &AuctionHandler{service: service}
}

func (h *AuctionHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /users", h.createUser)
	mux.HandleFunc("POST /items", h.createItem)
	mux.HandleFunc("GET /items", h.listItems)
	mux.HandleFunc("GET /items/{id}", h.getItem)
	mux.HandleFunc("POST /items/{id}/bid", h.placeBid)
	mux.HandleFunc("DELETE /items/{id}/bid/last", h.retractBid)
	mux.HandleFunc("GET /items/{id}/bids", h.getItemBids)
	mux.HandleFunc("POST /items/{id}/end", h.endAuction)
	mux.HandleFunc("GET /items/{id}/live", h.live)
	mux.HandleFunc("GET /stats", h.stats)
}

func (h *AuctionHandler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuctionHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}
	userID, err := h.service.CreateUser(strings.TrimSpace(req.Name), req.Balance)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user_id": userID})
}

func (h *AuctionHandler) createItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string   `json:"name"`
		CategoryPath []string `json:"category_path"`
		Description  string   `json:"description"`
		SellerID     int      `json:"seller_id"`
		StartPrice   float64  `json:"start_price"`
		StartTime    string   `json:"start_time"`
		EndTime      string   `json:"end_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_start_time", "start_time must be RFC3339")
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_end_time", "end_time must be RFC3339")
		return
	}

	itemID, err := h.service.CreateItem(
		req.Name, req.CategoryPath, req.Description,
		models.UserID(req.SellerID), req.StartPrice, startTime, endTime,
	)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"item_id": itemID})
}

func (h *AuctionHandler) listItems(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	if category == "" {
		writeError(w, http.StatusBadRequest, "missing_category", "category query param is required")
		return
	}
	items, err := h.service.BrowseCategory(category)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func parseItemID(r *http.Request) (models.ItemID, error) {
	raw := strings.TrimSpace(r.PathValue("id"))
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid_item_id")
	}
	return models.ItemID(id), nil
}

func (h *AuctionHandler) getItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_item_id", "item id must be positive integer")
		return
	}
	item, err := h.service.GetItem(itemID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *AuctionHandler) placeBid(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_item_id", "item id must be positive integer")
		return
	}
	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}
	bidID, err := h.service.PlaceBid(models.UserID(req.UserID), itemID, req.Amount)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"bid_id": bidID})
}

func (h *AuctionHandler) retractBid(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_item_id", "item id must be positive integer")
		return
	}
	var req struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}
	bidID, err := h.service.RetractLastBid(models.UserID(req.UserID), itemID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"retracted_bid_id": bidID})
}

func (h *AuctionHandler) getItemBids(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_item_id", "item id must be positive integer")
		return
	}
	bids, err := h.service.GetItemBids(itemID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bids)
}

func (h *AuctionHandler) endAuction(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_item_id", "item id must be positive integer")
		return
	}
	winner, err := h.service.EndAuction(itemID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"winner": winner})
}

func (h *AuctionHandler) stats(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.service.GetStats())
}

func (h *AuctionHandler) live(w http.ResponseWriter, r *http.Request) {
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_item_id", "item id must be positive integer")
		return
	}
	ch, err := h.service.SubscribeItem(itemID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	defer h.service.UnsubscribeItem(itemID, ch)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "stream_unsupported", "streaming unsupported")
		return
	}

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		case event := <-ch:
			payload, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: %s\n", event.Action)
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auction.ErrUserNotFound), errors.Is(err, auction.ErrItemNotFound), errors.Is(err, auction.ErrBidNotFound), errors.Is(err, auction.ErrCategoryNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, auction.ErrInvalidBidAmount), errors.Is(err, auction.ErrNoBidToRetract), errors.Is(err, auction.ErrInvalidUserInput), errors.Is(err, auction.ErrInvalidItemInput):
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, auction.ErrAuctionNotActive), errors.Is(err, auction.ErrSellerCannotBid), errors.Is(err, auction.ErrInsufficientBalance):
		writeError(w, http.StatusConflict, "invalid_state", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, statusCode int, code string, message string) {
	writeJSON(w, statusCode, map[string]string{
		"code":  code,
		"error": message,
	})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

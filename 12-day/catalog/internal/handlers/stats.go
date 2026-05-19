package handlers

import "net/http"

func (h *ProductHandler) Stats(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, h.store.Stats())
}

package ml

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	Client *Client
}

func (h *Handler) Classify(w http.ResponseWriter, r *http.Request) {
	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	result, err := h.Client.Classify(r.Context(), req.Text)
	if err != nil {
		http.Error(w, "ml error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

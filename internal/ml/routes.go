package ml

import "github.com/gorilla/mux"

func (h *Handler) RegisterRoutes(public, protected, csrf *mux.Router) {
	protected.HandleFunc("/classify", h.Classify).
		Methods("POST", "OPTIONS")
}

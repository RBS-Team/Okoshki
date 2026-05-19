package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (h *handler) RegisterRoutes(public, protected, csrfProtected *mux.Router) {
	public.HandleFunc("/login", h.Login).Methods(http.MethodPost, http.MethodOptions)
	public.HandleFunc("/guest/session", h.CreateGuestSession).Methods(http.MethodPost, http.MethodOptions)
	
	csrfProtected.HandleFunc("/csrf-token", h.GetCSRFToken).Methods(http.MethodGet, http.MethodOptions)
	csrfProtected.HandleFunc("/logout", h.Logout).Methods(http.MethodPost, http.MethodOptions)
}

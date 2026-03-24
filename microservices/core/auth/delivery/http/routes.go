package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (h *AuthHandler) RegisterRoutes(public, protected *mux.Router) {
	public.HandleFunc("/client/register", h.Register).Methods(http.MethodPost, http.MethodOptions)
	public.HandleFunc("/client/login", h.Login).Methods(http.MethodPost, http.MethodOptions)
	public.HandleFunc("/logout", h.Logout).Methods(http.MethodPost, http.MethodOptions)
	public.HandleFunc("/csrf-token", h.GetCSRFToken).Methods(http.MethodGet, http.MethodOptions)

}

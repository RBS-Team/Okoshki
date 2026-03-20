package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (h *Handler) RegisterRoutes(public *mux.Router) {
	public.HandleFunc("/categories", h.GetAllCategories).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/categories/{id}", h.GetCategoryByID).Methods(http.MethodGet, http.MethodOptions)

	public.HandleFunc("/masters", h.GetAllMasters).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/masters/{id}", h.GetMasterByID).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/masters", h.CreateMaster).Methods(http.MethodPost, http.MethodOptions)

	public.HandleFunc("/masters/{id}/services", h.GetServiceItemsByMasterID).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/masters/{id}/services", h.CreateServiceItem).Methods(http.MethodPost, http.MethodOptions)
}
package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/model"
)

func (h *Handler) RegisterRoutes(public, protected, csrfProtected *mux.Router) {
	public.HandleFunc("/categories", h.GetAllCategories).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/categories/{id}", h.GetCategoryByID).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/categories/{id}/masters", h.GetMastersByCategory).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/categories/{id}/services", h.GetServicesByCategory).Methods(http.MethodGet, http.MethodOptions)

	public.HandleFunc("/masters", h.GetAllMasters).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/masters/{id}", h.GetMasterByID).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/masters/{id}/services", h.GetServiceItemsByMasterID).Methods(http.MethodGet, http.MethodOptions)

	masterProtected := csrfProtected.PathPrefix("").Subrouter()
	masterProtected.Use(middleware.RequireRole(string(model.RoleMaster)))

	masterProtected.HandleFunc("/masters", h.CreateMaster).Methods(http.MethodPost, http.MethodOptions)
	masterProtected.HandleFunc("/masters/{id}/services", h.CreateServiceItem).Methods(http.MethodPost, http.MethodOptions)
}

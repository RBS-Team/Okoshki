package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/model"
)

func (h *Handler) RegisterRoutes(public, protected, csrfProtected *mux.Router) {
	public.HandleFunc("/master/register", h.RegisterMaster).Methods(http.MethodPost, http.MethodOptions)
	public.HandleFunc("/client/register", h.RegisterClient).Methods(http.MethodPost, http.MethodOptions)

	public.HandleFunc("/masters", h.GetAllMasters).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/masters/{id}", h.GetMasterByID).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/masters/{masterID}/portfolio", h.GetPortfolioPhotos).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/categories/{id}/masters", h.GetMastersByCategory).Methods(http.MethodGet, http.MethodOptions)
	
	masterProtected := csrfProtected.PathPrefix("").Subrouter()
	masterProtected.Use(middleware.RequireRole(string(model.RoleMaster)))

	masterProtected.HandleFunc("/masters/{masterID}/portfolio", h.UploadPortfolioPhotos).Methods(http.MethodPost, http.MethodOptions)
	masterProtected.HandleFunc("/masters/{masterID}/portfolio/{photoID}", h.DeletePortfolioPhoto).Methods(http.MethodDelete, http.MethodOptions)
}

package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/model"
)

func (h *handler) RegisterRoutes(public, protected, csrfProtected *mux.Router) {
	public.HandleFunc("/reviews/master/{masterID}", h.GetMasterReviews).Methods(http.MethodGet, http.MethodOptions)

	clientProtected := csrfProtected.PathPrefix("").Subrouter()
	clientProtected.Use(middleware.RequireRole(string(model.RoleClient)))

	clientProtected.HandleFunc("/reviews", h.CreateReview).Methods(http.MethodPost, http.MethodOptions)
	clientProtected.HandleFunc("/reviews/my", h.GetClientReviews).Methods(http.MethodGet, http.MethodOptions)

	protected.HandleFunc("/reviews/{id}", h.DeleteReview).Methods(http.MethodDelete, http.MethodOptions)
}

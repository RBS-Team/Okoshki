package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/model"
)

func (h *Handler) RegisterRoutes(public, protected, csrfProtected *mux.Router) {
	public.HandleFunc("/available-slots", h.GetAvailableSlots).Methods(http.MethodGet, http.MethodOptions)

	clientProtected := csrfProtected.PathPrefix("").Subrouter()
	clientProtected.Use(middleware.RequireRole(string(model.RoleClient)))

	clientProtected.HandleFunc("/appointments", h.CreateAppointment).Methods(http.MethodPost, http.MethodOptions)
}

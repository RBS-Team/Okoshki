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

	masterProtected := csrfProtected.PathPrefix("").Subrouter()
	masterProtected.Use(middleware.RequireRole(string(model.RoleMaster)))

	// Ручки клиента
	clientProtected.HandleFunc("/appointments", h.CreateAppointment).Methods(http.MethodPost, http.MethodOptions)
	clientProtected.HandleFunc("/appointments/my", h.GetMyAppointments).Methods(http.MethodGet, http.MethodOptions)
	clientProtected.HandleFunc("/appointments/{id}/cancel", h.CancelAppointment).Methods(http.MethodPatch, http.MethodOptions)

	// Ручки мастера
	masterProtected.HandleFunc("/master-appointments", h.GetMasterAppointments).Methods(http.MethodGet, http.MethodOptions)
	masterProtected.HandleFunc("/appointments/{id}/status", h.UpdateAppointmentStatus).Methods(http.MethodPatch, http.MethodOptions)
	masterProtected.HandleFunc("/appointments/block", h.CreateManualBlock).Methods(http.MethodPost, http.MethodOptions)
	masterProtected.HandleFunc("/appointments/block/{id}", h.DeleteManualBlock).Methods(http.MethodDelete, http.MethodOptions)
}

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
	public.HandleFunc("/categories/{id}/services", h.GetServicesByCategory).Methods(http.MethodGet, http.MethodOptions)

	public.HandleFunc("/masters/{id}/services", h.GetServiceItemsByMasterID).Methods(http.MethodGet, http.MethodOptions)

	masterProtected := csrfProtected.PathPrefix("").Subrouter()
	masterProtected.Use(middleware.RequireRole(string(model.RoleMaster)))

	masterProtected.HandleFunc("/masters/{id}/services", h.CreateServiceItem).Methods(http.MethodPost, http.MethodOptions)

	masterProtected.HandleFunc("/masters/{masterID}/working-hours", h.GetWorkingHours).Methods(http.MethodGet, http.MethodOptions)
	masterProtected.HandleFunc("/masters/{masterID}/working-hours", h.UpsertWorkingHours).Methods(http.MethodPut, http.MethodOptions)

	masterProtected.HandleFunc("/masters/{masterID}/schedule-exceptions", h.GetScheduleExceptions).Methods(http.MethodGet, http.MethodOptions)
	masterProtected.HandleFunc("/masters/{masterID}/schedule-exceptions", h.CreateScheduleException).Methods(http.MethodPost, http.MethodOptions)
	masterProtected.HandleFunc("/masters/{masterID}/schedule-exceptions/{id}", h.UpdateScheduleException).Methods(http.MethodPut, http.MethodOptions)
	masterProtected.HandleFunc("/masters/{masterID}/schedule-exceptions/{id}", h.DeleteScheduleException).Methods(http.MethodDelete, http.MethodOptions)
}

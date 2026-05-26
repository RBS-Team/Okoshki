package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/model"
)

// RegisterRoutes регистрирует маршруты каталога.
// masterCtx — middleware, который резолвит master_id из JWT и кладёт в контекст.
// Применяется к эндпоинтам /me/..., где мастер действует над собственными ресурсами.
func (h *handler) RegisterRoutes(public, protected, csrfProtected *mux.Router, masterCtx mux.MiddlewareFunc) {
	public.HandleFunc("/categories", h.GetAllCategories).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/categories/{id}", h.GetCategoryByID).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/categories/{id}/services", h.GetServicesByCategory).Methods(http.MethodGet, http.MethodOptions)

	public.HandleFunc("/masters/{id}/services", h.GetServiceItemsByMasterID).Methods(http.MethodGet, http.MethodOptions)
	public.HandleFunc("/masters/{masterID}/work-intervals", h.ListWorkIntervals).Methods(http.MethodGet, http.MethodOptions)

	adminProtected := csrfProtected.PathPrefix("").Subrouter()
	adminProtected.Use(middleware.RequireRole(string(model.RoleAdmin)))
	adminProtected.HandleFunc("/categories/{id}/avatar", h.UploadCategoryAvatar).Methods(http.MethodPut, http.MethodOptions)

	masterProtected := csrfProtected.PathPrefix("").Subrouter()
	masterProtected.Use(middleware.RequireRole(string(model.RoleMaster)))

	masterProtected.HandleFunc("/masters/{id}/services", h.CreateServiceItem).Methods(http.MethodPost, http.MethodOptions)

	// /me/... — собственные ресурсы мастера. master_id берётся из контекста, не из URL.
	me := masterProtected.PathPrefix("/me").Subrouter()
	me.Use(masterCtx)

	me.HandleFunc("/settings", h.GetMasterSettings).Methods(http.MethodGet, http.MethodOptions)
	me.HandleFunc("/settings", h.UpsertMasterSettings).Methods(http.MethodPut, http.MethodOptions)

	me.HandleFunc("/work-intervals", h.CreateWorkInterval).Methods(http.MethodPost, http.MethodOptions)
	me.HandleFunc("/work-intervals", h.ReplaceWorkIntervalsForDate).Methods(http.MethodPut, http.MethodOptions)
	me.HandleFunc("/work-intervals/{intervalID}", h.DeleteWorkInterval).Methods(http.MethodDelete, http.MethodOptions)
}

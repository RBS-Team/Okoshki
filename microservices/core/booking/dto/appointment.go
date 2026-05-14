package dto

import "time"

//go:generate easyjson $GOFILE

//easyjson:json
type GetAvailableSlotsResponse struct {
	Slots map[string][]string `json:"slots"`
}

//easyjson:json
type CreateAppointmentRequest struct {
	ServiceID     string  `json:"service_id"`
	StartAt       string  `json:"start_at"`
	ClientComment *string `json:"client_comment,omitempty"`
}

//easyjson:json
type AppointmentResponse struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	MasterID      string    `json:"master_id"`
	ServiceID     string    `json:"service_id"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	Status        string    `json:"status"`
	ClientComment *string   `json:"client_comment,omitempty"`
}

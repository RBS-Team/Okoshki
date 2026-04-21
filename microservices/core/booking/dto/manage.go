package dto

import "time"

//go:generate easyjson $GOFILE

//easyjson:json
type ClientAppointmentView struct {
	ID            string    `json:"id"`
	MasterID      string    `json:"master_id"`
	MasterName    string    `json:"master_name"`
	MasterAvatar  *string   `json:"master_avatar,omitempty"`
	MasterLat     *float64  `json:"master_lat,omitempty"`
	MasterLon     *float64  `json:"master_lon,omitempty"`
	ServiceID     string    `json:"service_id"`
	ServiceTitle  string    `json:"service_title"`
	Price         float64   `json:"price"`
	Duration      int       `json:"duration_minutes"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	Status        string    `json:"status"`
	ClientComment *string   `json:"client_comment,omitempty"`
	MasterNote    *string   `json:"master_note,omitempty"`
}

//easyjson:json
type MasterAppointmentView struct {
	ID            string    `json:"id"`
	ClientID      *string   `json:"client_id,omitempty"`
	ClientEmail   *string   `json:"client_email,omitempty"`
	ClientAvatar  *string   `json:"client_avatar,omitempty"`
	ServiceID     *string   `json:"service_id,omitempty"`
	ServiceTitle  *string   `json:"service_title,omitempty"`
	Price         *float64  `json:"price,omitempty"`
	Duration      *int      `json:"duration_minutes,omitempty"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	Status        string    `json:"status"`
	IsManualBlock bool      `json:"is_manual_block"`
	ClientComment *string   `json:"client_comment,omitempty"`
	MasterNote    *string   `json:"master_note,omitempty"`
}

//easyjson:json
type UpdateAppointmentStatusRequest struct {
	Status     string  `json:"status"`
	MasterNote *string `json:"master_note,omitempty"`
}

//easyjson:json
type CreateManualBlockRequest struct {
	StartAt string  `json:"start_at"`
	EndAt   string  `json:"end_at"`
	Note    *string `json:"note,omitempty"`
}

//easyjson:json
type CreateManualBlockResponse struct {
	ID      string    `json:"id"`
	StartAt time.Time `json:"start_at"`
	EndAt   time.Time `json:"end_at"`
	Note    *string   `json:"note,omitempty"`
}

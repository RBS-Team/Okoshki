package dto

import "time"

//go:generate easyjson $GOFILE

//easyjson:json
type CreateReviewRequest struct {
	AppointmentID string `json:"appointment_id"`
	Rating        int    `json:"rating"`
	Comment       string `json:"comment"`
}

//easyjson:json
type ReviewResponse struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	MasterID      string    `json:"master_id"`
	AppointmentID string    `json:"appointment_id"`
	Rating        int       `json:"rating"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}

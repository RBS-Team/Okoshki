package dto

//go:generate easyjson $GOFILE

//easyjson:json
type CsrfResponse struct {
	Token string `json:"csrf_token"`
}

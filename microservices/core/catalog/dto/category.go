package dto

//go:generate easyjson $GOFILE

//easyjson:json
type Category struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	MastersCount int     `json:"masters_count"`
}

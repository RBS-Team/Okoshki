package dto

//go:generate easyjson $GOFILE

//easyjson:json
type Client struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     string  `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

//easyjson:json
type RegisterClientRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

//easyjson:json
type RegisterClientResponse struct {
	UserID   string `json:"user_id"`
	ClientID string `json:"client_id"`
	Role     string `json:"role"`
}

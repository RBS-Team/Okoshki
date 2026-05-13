package dto

//go:generate easyjson $GOFILE

//easyjson:json
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

//easyjson:json
type LoginResponse struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

//easyjson:json
type UserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

//easyjson:json
type GuestSessionResponse struct {
	GuestID string `json:"guest_id"`
	Role string `json:"role"`
}

package dto

//go:generate easyjson $GOFILE

//easyjson:json
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

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
type RegisterResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
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
}

//easyjson:json
type RegisterMasterRequest struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Name     string   `json:"name"`
	Bio      *string  `json:"bio,omitempty"`
	Timezone string   `json:"timezone,omitempty"`
	Lat      *float64 `json:"lat,omitempty"`
	Lon      *float64 `json:"lon,omitempty"`
}

//easyjson:json
type RegisterMasterResponse struct {
	UserID   string `json:"user_id"`
	MasterID string `json:"master_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

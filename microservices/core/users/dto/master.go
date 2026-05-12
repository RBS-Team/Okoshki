package dto

//go:generate easyjson $GOFILE

//easyjson:json
type Master struct {
	ID          string   `json:"id"`
	UserID      string   `json:"user_id"`
	CategoryID  string   `json:"category_id"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Address     string   `json:"address"`
	City        string   `json:"city"`
	Bio         *string  `json:"bio,omitempty"`
	AvatarURL   *string  `json:"avatar_url,omitempty"`
	Timezone    string   `json:"timezone"`
	Lat         *float64 `json:"lat,omitempty"`
	Lon         *float64 `json:"lon,omitempty"`
	Rating      float64  `json:"rating"`
	ReviewCount int      `json:"review_count"`
}

//easyjson:json
type CreateMasterRequest struct {
	CategoryID string   `json:"category_id"`
	FirstName  string   `json:"first_name"`
	LastName   string   `json:"last_name"`
	Address    string   `json:"address"`
	City       string   `json:"city"`
	Bio        *string  `json:"bio,omitempty"`
	Timezone   string   `json:"timezone,omitempty"`
	Lat        *float64 `json:"lat,omitempty"`
	Lon        *float64 `json:"lon,omitempty"`
}

//easyjson:json
type RegisterMasterRequest struct {
	Email      string   `json:"email"`
	Password   string   `json:"password"`
	CategoryID string   `json:"category_id"`
	FirstName  string   `json:"first_name"`
	LastName   string   `json:"last_name"`
	Address    string   `json:"address"`
	City       string   `json:"city"`
	Bio        *string  `json:"bio,omitempty"`
	Timezone   string   `json:"timezone,omitempty"`
	Lat        *float64 `json:"lat,omitempty"`
	Lon        *float64 `json:"lon,omitempty"`
}

//easyjson:json
type RegisterMasterResponse struct {
	UserID   string `json:"user_id"`
	MasterID string `json:"master_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

//easyjson:json
type UpdateMasterRequest struct {
	CategoryID *string  `json:"category_id,omitempty"`
	FirstName  *string  `json:"first_name,omitempty"`
	LastName   *string  `json:"last_name,omitempty"`
	Address    *string  `json:"address,omitempty"`
	City       *string  `json:"city,omitempty"`
	Bio        *string  `json:"bio,omitempty"`
	Timezone   *string  `json:"timezone,omitempty"`
	Lat        *float64 `json:"lat,omitempty"`
	Lon        *float64 `json:"lon,omitempty"`
}

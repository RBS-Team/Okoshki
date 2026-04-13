package dto

//go:generate easyjson $GOFILE

//easyjson:json
type Master struct {
	ID          string   `json:"id"`
	UserID      string   `json:"user_id"`
	Name        string   `json:"name"`
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
	Name     string   `json:"name"`
	Bio      *string  `json:"bio,omitempty"`
	Timezone string   `json:"timezone,omitempty"`
	Lat      *float64 `json:"lat,omitempty"`
	Lon      *float64 `json:"lon,omitempty"`
}

//easyjson:json
type UpdateMasterRequest struct {
	Name     *string  `json:"name,omitempty"`
	Bio      *string  `json:"bio,omitempty"`
	Timezone *string  `json:"timezone,omitempty"`
	Lat      *float64 `json:"lat,omitempty"`
	Lon      *float64 `json:"lon,omitempty"`
}

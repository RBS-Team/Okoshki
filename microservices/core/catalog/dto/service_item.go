package dto

//go:generate easyjson $GOFILE

//easyjson:json
type ServiceItem struct {
	ID                  string  `json:"id"`
	MasterID            string  `json:"master_id"`
	CategoryID          string  `json:"category_id"`
	Title               string  `json:"title"`
	Address             string  `json:"address"`
	Description         *string `json:"description,omitempty"`
	Price               float64 `json:"price"`
	DurationMinutes     int     `json:"duration_minutes"`
	BufferBeforeMinutes int     `json:"buffer_before_minutes"`
	BufferAfterMinutes  int     `json:"buffer_after_minutes"`
	IsActive            bool    `json:"is_active"`
	IsAutoConfirm       bool    `json:"is_auto_confirm"`
}

//easyjson:json
type CreateServiceItemRequest struct {
	CategoryID          string  `json:"category_id"`
	Title               string  `json:"title"`
	Address             string  `json:"address"`
	Description         *string `json:"description,omitempty"`
	Price               float64 `json:"price"`
	DurationMinutes     int     `json:"duration_minutes"`
	BufferBeforeMinutes int     `json:"buffer_before_minutes,omitempty"`
	BufferAfterMinutes  int     `json:"buffer_after_minutes,omitempty"`
	IsAutoConfirm       *bool   `json:"is_auto_confirm,omitempty"`
}

//easyjson:json
type ServiceWithMaster struct {
	ID                  string   `json:"id"`
	CategoryID          string   `json:"category_id"`
	Title               string   `json:"title"`
	Description         *string  `json:"description,omitempty"`
	Price               float64  `json:"price"`
	DurationMinutes     int      `json:"duration_minutes"`
	BufferBeforeMinutes int      `json:"buffer_before_minutes"`
	BufferAfterMinutes  int      `json:"buffer_after_minutes"`
	IsActive            bool     `json:"is_active"`
	IsAutoConfirm       bool     `json:"is_auto_confirm"`
	
	MasterID            string   `json:"master_id"`
	FirstName           string   `json:"first_name"`
	LastName            string   `json:"last_name"`
	Address             string   `json:"address"`
	City                string   `json:"city"`
	Bio                 *string  `json:"bio,omitempty"`
	AvatarURL           *string  `json:"avatar_url,omitempty"`
	Timezone            string   `json:"timezone"`
	Lat                 *float64 `json:"lat,omitempty"`
	Lon                 *float64 `json:"lon,omitempty"`
	Rating              float64  `json:"rating"`
	ReviewCount         int      `json:"review_count"`
}

package dto

//go:generate easyjson $GOFILE

//easyjson:json
type ServiceItem struct {
	ID                  string  `json:"id"`
	MasterID            string  `json:"master_id"`
	CategoryID          string  `json:"category_id"`
	Title               string  `json:"title"`
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
	Description         *string `json:"description,omitempty"`
	Price               float64 `json:"price"`
	DurationMinutes     int     `json:"duration_minutes"`
	BufferBeforeMinutes int     `json:"buffer_before_minutes,omitempty"`
	BufferAfterMinutes  int     `json:"buffer_after_minutes,omitempty"`
	IsAutoConfirm       *bool   `json:"is_auto_confirm,omitempty"`
}

//easyjson:json
type ServiceWithMaster struct {
	ID                  string  `json:"id"`
	CategoryID          string  `json:"category_id"`
	Title               string  `json:"title"`
	Description         *string `json:"description,omitempty"`
	Price               float64 `json:"price"`
	DurationMinutes     int     `json:"duration_minutes"`
	BufferBeforeMinutes int     `json:"buffer_before_minutes"`
	BufferAfterMinutes  int     `json:"buffer_after_minutes"`
	IsActive            bool    `json:"is_active"`
	IsAutoConfirm       bool    `json:"is_auto_confirm"`
	Master              Master  `json:"master"`
}

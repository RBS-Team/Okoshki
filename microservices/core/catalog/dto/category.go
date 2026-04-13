package dto

//go:generate easyjson $GOFILE

//easyjson:json
type Category struct {
	ID          string      `json:"id"`
	ParentID    *string     `json:"parent_id,omitempty"`
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Children    []*Category `json:"children,omitempty"`
}

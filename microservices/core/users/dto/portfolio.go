package dto

//go:generate easyjson $GOFILE

import "io"

type FileUpload struct {
	Reader      io.Reader
	Size        int64
	ContentType string
	Name        string
}

//easyjson:json
type PortfolioPhoto struct {
	ID       string `json:"id"`
	MasterID string `json:"master_id"`
	URL      string `json:"url"`
}

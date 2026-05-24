package ml

type Request struct {
	Text string `json:"text"`
}

type Response struct {
	RecommendedService string `json:"recommended_service"`
}

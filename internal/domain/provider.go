package domain

import "context"

type SearchRequest struct {
	From string
	To   string
	Date string
}

type FlightResult struct {
	Provider string `json:"provider,omitempty"`
	ID       string `json:"id"`
	From     string `json:"from"`
	To       string `json:"to"`
	Date     string `json:"date"`
	Price    int    `json:"price"`
	Currency string `json:"currency"`
	Airline  string `json:"airline"`
}

type ProviderSearchResponse struct {
	Provider string         `json:"provider"`
	Results  []FlightResult `json:"results"`
}

type Provider interface {
	Name() string
	Search(ctx context.Context, req SearchRequest) (*ProviderSearchResponse, error)
}

package domain

import "context"

type HotelSearchRequest struct {
	City     string
	CheckIn  string
	CheckOut string
}

type HotelResult struct {
	Provider      string `json:"provider,omitempty"`
	ID            string `json:"id"`
	City          string `json:"city"`
	CheckIn       string `json:"checkin"`
	CheckOut      string `json:"checkout"`
	Name          string `json:"name"`
	PricePerNight int    `json:"price_per_night"`
	Currency      string `json:"currency"`
}

type HotelProviderSearchResponse struct {
	Provider string        `json:"provider"`
	Results  []HotelResult `json:"results"`
}

type HotelProvider interface {
	Name() string
	Search(ctx context.Context, req HotelSearchRequest) (*HotelProviderSearchResponse, error)
}
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yourusername/goaggregator/internal/domain"
)

type HTTPHotelProvider struct {
	name    string
	baseURL string
	client  *http.Client
}

func NewHTTPHotelProvider(name, baseURL string, timeout time.Duration) *HTTPHotelProvider {
	return &HTTPHotelProvider{
		name:    name,
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func NewHTTPHotelProviders(hosts []string, timeout time.Duration) []domain.HotelProvider {
	providers := make([]domain.HotelProvider, 0, len(hosts))
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}
		providers = append(providers, NewHTTPHotelProvider(providerName(host), host, timeout))
	}
	return providers
}

func (p *HTTPHotelProvider) Name() string {
	return p.name
}

func (p *HTTPHotelProvider) Search(ctx context.Context, req domain.HotelSearchRequest) (*domain.HotelProviderSearchResponse, error) {
	endpoint, err := url.Parse(p.baseURL + "/search/hotels")
	if err != nil {
		return nil, fmt.Errorf("parse provider url %q: %w", p.baseURL, err)
	}

	query := endpoint.Query()
	if req.City != "" {
		query.Set("city", req.City)
	}
	if req.CheckIn != "" {
		query.Set("checkin", req.CheckIn)
	}
	if req.CheckOut != "" {
		query.Set("checkout", req.CheckOut)
	}
	endpoint.RawQuery = query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create provider request %q: %w", p.name, err)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call provider %q: %w", p.name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("provider %q returned status %d", p.name, resp.StatusCode)
	}

	var searchResp domain.HotelProviderSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode provider %q response: %w", p.name, err)
	}
	if searchResp.Provider == "" {
		searchResp.Provider = p.name
	}
	return &searchResp, nil
}
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

type HTTPProvider struct {
	name    string
	baseURL string
	client  *http.Client
}

func NewHTTPProvider(name, baseURL string, timeout time.Duration) *HTTPProvider {
	return &HTTPProvider{
		name:    name,
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func NewHTTPProviders(hosts []string, timeout time.Duration) []domain.Provider {
	providers := make([]domain.Provider, 0, len(hosts))
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}
		providers = append(providers, NewHTTPProvider(host, host, timeout))
	}
	return providers
}

func (p *HTTPProvider) Name() string {
	return p.name
}

func (p *HTTPProvider) Search(ctx context.Context, req domain.SearchRequest) (*domain.ProviderSearchResponse, error) {
	endpoint, err := url.Parse(p.baseURL + "/search")
	if err != nil {
		return nil, fmt.Errorf("parse provider url %q: %w", p.baseURL, err)
	}

	query := endpoint.Query()
	if req.From != "" {
		query.Set("from", req.From)
	}
	if req.To != "" {
		query.Set("to", req.To)
	}
	if req.Date != "" {
		query.Set("date", req.Date)
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

	var searchResp domain.ProviderSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode provider %q response: %w", p.name, err)
	}
	if searchResp.Provider == "" {
		searchResp.Provider = p.name
	}
	return &searchResp, nil
}

func providerName(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return strings.TrimRight(rawURL, "/")
	}
	return parsed.Hostname()
}

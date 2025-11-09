package validation

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// HTTPValidator validates domains using HTTP requests
type HTTPValidator struct {
	client *http.Client
}

// NewHTTPValidator creates a new HTTP validator
func NewHTTPValidator() *HTTPValidator {
	return &HTTPValidator{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
				},
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  true,
				DisableKeepAlives:   true,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Allow up to 10 redirects
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
	}
}

// Validate checks if a domain is reachable via HTTP/HTTPS
func (h *HTTPValidator) Validate(ctx context.Context, domain string) (*ValidationResult, error) {
	result := &ValidationResult{
		Domain:  domain,
		Details: make(map[string]interface{}),
	}

	// Try HTTPS first
	httpsURL := fmt.Sprintf("https://%s", domain)
	httpsResp, httpsErr := h.makeRequest(ctx, httpsURL)
	if httpsErr == nil {
		result.Valid = true
		result.Message = fmt.Sprintf("Domain is reachable via HTTPS (status: %d)", httpsResp.StatusCode)
		result.Details["protocol"] = "https"
		result.Details["status_code"] = httpsResp.StatusCode
		result.Details["final_url"] = httpsResp.Request.URL.String()
		return result, nil
	}

	// Try HTTP if HTTPS fails
	httpURL := fmt.Sprintf("http://%s", domain)
	httpResp, httpErr := h.makeRequest(ctx, httpURL)
	if httpErr == nil {
		result.Valid = true
		result.Message = fmt.Sprintf("Domain is reachable via HTTP (status: %d)", httpResp.StatusCode)
		result.Details["protocol"] = "http"
		result.Details["status_code"] = httpResp.StatusCode
		result.Details["final_url"] = httpResp.Request.URL.String()
		return result, nil
	}

	// Both failed
	result.Valid = false
	result.Message = "Domain is not reachable via HTTP or HTTPS"
	result.Details["errors"] = map[string]string{
		"https": formatError(httpsErr),
		"http":  formatError(httpErr),
	}

	return result, nil
}

func (h *HTTPValidator) makeRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Kestrel-ThreatIntel/1.0")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx as successful
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return resp, nil
	}

	return resp, fmt.Errorf("HTTP status %d", resp.StatusCode)
}

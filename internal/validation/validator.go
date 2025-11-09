package validation

import (
	"context"
	"fmt"
)

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	Valid   bool
	Domain  string
	Message string
	Details map[string]interface{}
}

// ValidationMode represents the type of validation to perform
type ValidationMode int

const (
	ValidationNone ValidationMode = iota
	ValidationDNS
	ValidationHTTP
	ValidationFull
)

// Validator validates domains for threat intelligence feeds
type Validator struct {
	dnsValidator  *DNSValidator
	httpValidator *HTTPValidator
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		dnsValidator:  NewDNSValidator(),
		httpValidator: NewHTTPValidator(),
	}
}

// Validate validates a domain based on the specified mode
func (v *Validator) Validate(ctx context.Context, domain string, mode ValidationMode) (*ValidationResult, error) {
	switch mode {
	case ValidationNone:
		return &ValidationResult{
			Valid:   true,
			Domain:  domain,
			Message: "No validation performed",
		}, nil

	case ValidationDNS:
		return v.dnsValidator.Validate(ctx, domain)

	case ValidationHTTP:
		return v.httpValidator.Validate(ctx, domain)

	case ValidationFull:
		return v.validateFull(ctx, domain)

	default:
		return nil, fmt.Errorf("unknown validation mode: %d", mode)
	}
}

// validateFull performs both DNS and HTTP validation
func (v *Validator) validateFull(ctx context.Context, domain string) (*ValidationResult, error) {
	// First check DNS
	dnsResult, err := v.dnsValidator.Validate(ctx, domain)
	if err != nil {
		return nil, err
	}

	if !dnsResult.Valid {
		return dnsResult, nil
	}

	// Then check HTTP
	httpResult, err := v.httpValidator.Validate(ctx, domain)
	if err != nil {
		return nil, err
	}

	// Combine results
	result := &ValidationResult{
		Valid:  httpResult.Valid,
		Domain: domain,
		Details: map[string]interface{}{
			"dns":  dnsResult.Details,
			"http": httpResult.Details,
		},
	}

	if httpResult.Valid {
		result.Message = "Domain passed full validation (DNS + HTTP)"
	} else {
		result.Message = fmt.Sprintf("Domain failed HTTP validation: %s", httpResult.Message)
	}

	return result, nil
}

// ParseValidationMode parses a string into a ValidationMode
func ParseValidationMode(mode string) ValidationMode {
	switch mode {
	case "dns", "dns_validate":
		return ValidationDNS
	case "http", "http_validate":
		return ValidationHTTP
	case "full", "full_validate":
		return ValidationFull
	default:
		return ValidationNone
	}
}

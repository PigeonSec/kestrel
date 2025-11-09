package validation

import (
	"context"
	"fmt"
	"net"
)

// DNSValidator validates domains using DNS lookups
type DNSValidator struct {
	resolver *net.Resolver
}

// NewDNSValidator creates a new DNS validator
func NewDNSValidator() *DNSValidator {
	return &DNSValidator{
		resolver: &net.Resolver{
			PreferGo: true,
		},
	}
}

// Validate checks if a domain has valid DNS records (A, AAAA, or CNAME)
func (d *DNSValidator) Validate(ctx context.Context, domain string) (*ValidationResult, error) {
	result := &ValidationResult{
		Domain:  domain,
		Details: make(map[string]interface{}),
	}

	// Check for A records
	aRecords, aErr := d.resolver.LookupIP(ctx, "ip4", domain)
	if aErr == nil && len(aRecords) > 0 {
		result.Valid = true
		result.Message = fmt.Sprintf("Domain has %d A record(s)", len(aRecords))
		result.Details["a_records"] = formatIPs(aRecords)
		return result, nil
	}

	// Check for AAAA records
	aaaaRecords, aaaaErr := d.resolver.LookupIP(ctx, "ip6", domain)
	if aaaaErr == nil && len(aaaaRecords) > 0 {
		result.Valid = true
		result.Message = fmt.Sprintf("Domain has %d AAAA record(s)", len(aaaaRecords))
		result.Details["aaaa_records"] = formatIPs(aaaaRecords)
		return result, nil
	}

	// Check for CNAME records
	cname, cnameErr := d.resolver.LookupCNAME(ctx, domain)
	if cnameErr == nil && cname != "" && cname != domain+"." {
		result.Valid = true
		result.Message = "Domain has CNAME record"
		result.Details["cname"] = cname
		return result, nil
	}

	// No valid records found
	result.Valid = false
	result.Message = "No valid DNS records found (no A, AAAA, or CNAME)"
	result.Details["errors"] = map[string]string{
		"a_record":     formatError(aErr),
		"aaaa_record":  formatError(aaaaErr),
		"cname_record": formatError(cnameErr),
	}

	return result, nil
}

func formatIPs(ips []net.IP) []string {
	result := make([]string, len(ips))
	for i, ip := range ips {
		result[i] = ip.String()
	}
	return result
}

func formatError(err error) string {
	if err == nil {
		return "no error"
	}
	return err.Error()
}

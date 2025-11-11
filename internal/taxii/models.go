package taxii

import "time"

// TAXII 2.1 Data Models
// Spec: https://docs.oasis-open.org/cti/taxii/v2.1/taxii-v2.1.html

// Discovery represents the TAXII discovery response
type Discovery struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Contact     string   `json:"contact,omitempty"`
	Default     string   `json:"default,omitempty"`
	APIRoots    []string `json:"api_roots"`
}

// APIRoot represents a TAXII API root
type APIRoot struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Versions         []string `json:"versions"`
	MaxContentLength int64    `json:"max_content_length"`
}

// Collections represents a list of TAXII collections
type Collections struct {
	Collections []Collection `json:"collections"`
}

// Collection represents a TAXII collection
type Collection struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	CanRead     bool     `json:"can_read"`
	CanWrite    bool     `json:"can_write"`
	MediaTypes  []string `json:"media_types,omitempty"`
}

// Envelope wraps STIX objects in a TAXII response
type Envelope struct {
	More    bool   `json:"more"`
	Next    string `json:"next,omitempty"`
	Objects []any  `json:"objects"`
}

// ManifestRecord represents a manifest entry
type ManifestRecord struct {
	ID         string `json:"id"`
	DateAdded  string `json:"date_added"`
	Version    string `json:"version"`
	MediaType  string `json:"media_type,omitempty"`
}

// Manifest represents a collection manifest
type Manifest struct {
	More    bool             `json:"more"`
	Next    string           `json:"next,omitempty"`
	Objects []ManifestRecord `json:"objects"`
}

// Error represents a TAXII error response
type Error struct {
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	ErrorID     string            `json:"error_id,omitempty"`
	ErrorCode   string            `json:"error_code,omitempty"`
	HTTPStatus  int               `json:"http_status,omitempty"`
	ExternalDetails []ExternalDetail `json:"external_details,omitempty"`
	Details     map[string]any    `json:"details,omitempty"`
}

// ExternalDetail provides external error information
type ExternalDetail struct {
	SourceName  string `json:"source_name"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

// Status represents the status of an asynchronous request
type Status struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	RequestTimestamp string `json:"request_timestamp,omitempty"`
	TotalCount       int    `json:"total_count,omitempty"`
	SuccessCount     int    `json:"success_count,omitempty"`
	FailureCount     int    `json:"failure_count,omitempty"`
	PendingCount     int    `json:"pending_count,omitempty"`
}

// Versions represents TAXII server versions
type Versions struct {
	Versions []string `json:"versions"`
}

// FilterParams represents query parameters for filtering
type FilterParams struct {
	AddedAfter string   // RFC3339 timestamp
	Limit      int      // Max objects to return
	Next       string   // Pagination token
	Match      []string // Object IDs to match
	Type       []string // Object types to filter
	Version    string   // Object version
}

// Constants for TAXII 2.1
const (
	MediaTypeTAXII      = "application/taxii+json;version=2.1"
	MediaTypeSTIX       = "application/stix+json;version=2.1"
	TAXIIVersion        = "2.1"
	MaxContentLength    = 104857600 // 100MB
	DefaultLimit        = 100
	MaxLimit            = 1000
)

// Helper to format timestamps
func FormatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

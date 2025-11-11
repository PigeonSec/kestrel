package stix

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// IOCData represents the internal IOC data structure that needs to be converted to STIX
type IOCData struct {
	// IOC Value (one required)
	Domain   string
	IP       string
	URL      string
	Hash     string
	HashType string // md5, sha1, sha256
	Email    string

	// Metadata
	Category     string
	Comment      string
	Feed         string
	Source       string
	Confidence   int
	TLP          string
	FirstSeen    time.Time
	LastSeen     time.Time
	Organization string

	// Enrichment
	ThreatActor string
	Malware     string
	Campaign    string
	Tags        []string
}

// ConversionOptions controls how STIX objects are generated
type ConversionOptions struct {
	IncludeIdentity    bool
	IncludeObservedData bool
	IncludeRelationships bool
	StixID             string // If provided, reuse this STIX ID instead of generating new
}

// Converter handles conversion from internal models to STIX 2.1
type Converter struct {
	organizationIdentityID string
}

// NewConverter creates a new STIX converter
func NewConverter() *Converter {
	return &Converter{
		organizationIdentityID: "identity--" + uuid.New().String(),
	}
}

// ConvertDomainToSTIX converts a domain IOC to a STIX bundle
func (c *Converter) ConvertDomainToSTIX(ioc IOCData, opts ConversionOptions) Bundle {
	now := FormatTimestamp(time.Now().UTC())

	// Generate or reuse STIX ID
	indicatorID := opts.StixID
	if indicatorID == "" {
		indicatorID = "indicator--" + uuid.New().String()
	}

	// Create the indicator
	indicator := c.createDomainIndicator(ioc, indicatorID, now)

	// Start building bundle objects
	objects := []any{indicator}

	// Add identity if requested
	if opts.IncludeIdentity {
		identity := c.createIdentity(ioc, now)
		objects = append(objects, identity)
		indicator.CreatedByRef = identity.ID
	}

	// Add observed data if requested
	if opts.IncludeObservedData {
		observedData := c.createObservedData(ioc, now)
		objects = append(objects, observedData)

		// Add relationship if we have both indicator and observed data
		if opts.IncludeRelationships {
			rel := c.createRelationship(indicatorID, observedData.ID, "based-on", now)
			objects = append(objects, rel)
		}
	}

	// Create bundle
	bundleID := "bundle--" + uuid.New().String()
	return Bundle{
		Type:    "bundle",
		ID:      bundleID,
		Objects: objects,
	}
}

// createDomainIndicator creates a STIX indicator for any IOC type
func (c *Converter) createDomainIndicator(ioc IOCData, indicatorID, timestamp string) Indicator {
	// Build STIX pattern based on IOC type
	pattern, iocType := c.buildSTIXPattern(ioc)

	// Map category to indicator types
	indicatorTypes := c.mapCategoryToIndicatorTypes(ioc.Category)

	// Build description
	description := ioc.Comment
	if description == "" {
		description = fmt.Sprintf("Malicious %s detected in category: %s", iocType, ioc.Category)
	}

	// Get TLP marking
	tlpRef := GetTLPMarkingRef(ioc.TLP)

	// Build name
	name := c.buildIndicatorName(ioc, iocType)

	indicator := Indicator{
		Type:             "indicator",
		SpecVersion:      "2.1",
		ID:               indicatorID,
		Created:          FormatTimestamp(ioc.FirstSeen),
		Modified:         FormatTimestamp(ioc.LastSeen),
		Name:             name,
		Description:      description,
		IndicatorTypes:   indicatorTypes,
		Pattern:          pattern,
		PatternType:      "stix",
		PatternVersion:   "2.1",
		ValidFrom:        FormatTimestamp(ioc.FirstSeen),
		Confidence:       ioc.Confidence,
		Labels:           c.buildLabels(ioc),
		ObjectMarkingRefs: []string{tlpRef},
	}

	// Add external references if source is provided
	if ioc.Source != "" {
		indicator.ExternalReferences = []ExternalReference{
			{
				SourceName:  ioc.Source,
				Description: fmt.Sprintf("IOC from %s feed", ioc.Feed),
			},
		}
	}

	return indicator
}

// createIdentity creates a STIX identity object
func (c *Converter) createIdentity(ioc IOCData, timestamp string) Identity {
	orgName := ioc.Organization
	if orgName == "" {
		orgName = "Kestrel CTI"
	}

	return Identity{
		Type:          "identity",
		SpecVersion:   "2.1",
		ID:            c.organizationIdentityID,
		Created:       timestamp,
		Modified:      timestamp,
		Name:          orgName,
		IdentityClass: "organization",
		Sectors:       []string{"technology"},
		Description:   "Threat Intelligence Provider",
	}
}

// createObservedData creates a STIX observed-data object
func (c *Converter) createObservedData(ioc IOCData, timestamp string) ObservedData {
	observedDataID := "observed-data--" + uuid.New().String()

	return ObservedData{
		Type:           "observed-data",
		SpecVersion:    "2.1",
		ID:             observedDataID,
		Created:        timestamp,
		Modified:       timestamp,
		FirstObserved:  FormatTimestamp(ioc.FirstSeen),
		LastObserved:   FormatTimestamp(ioc.LastSeen),
		NumberObserved: 1,
		Objects: map[string]ObservableObject{
			"0": {
				Type:  "domain-name",
				Value: ioc.Domain,
			},
		},
	}
}

// createRelationship creates a STIX relationship object
func (c *Converter) createRelationship(sourceRef, targetRef, relType, timestamp string) Relationship {
	return Relationship{
		Type:             "relationship",
		SpecVersion:      "2.1",
		ID:               "relationship--" + uuid.New().String(),
		Created:          timestamp,
		Modified:         timestamp,
		RelationshipType: relType,
		SourceRef:        sourceRef,
		TargetRef:        targetRef,
		Description:      fmt.Sprintf("Indicator %s %s", relType, targetRef),
	}
}

// mapCategoryToIndicatorTypes maps internal categories to STIX indicator types
func (c *Converter) mapCategoryToIndicatorTypes(category string) []string {
	mapping := map[string][]string{
		"Malware":           {"malicious-activity"},
		"Phishing":          {"malicious-activity", "phishing"},
		"C2":                {"malicious-activity", "command-and-control"},
		"Botnet":            {"malicious-activity", "botnet"},
		"APT":               {"malicious-activity", "attribution"},
		"Exploit":           {"malicious-activity"},
		"Ransomware":        {"malicious-activity"},
		"Cryptomining":      {"malicious-activity"},
		"Spam":              {"anomalous-activity"},
		"Suspicious":        {"anomalous-activity"},
		"Benign":            {"benign"},
	}

	if types, ok := mapping[category]; ok {
		return types
	}

	// Default
	return []string{"malicious-activity"}
}

// buildLabels creates labels array from IOC data
func (c *Converter) buildLabels(ioc IOCData) []string {
	labels := []string{
		ioc.Category,
		ioc.Feed,
	}

	if ioc.Source != "" {
		labels = append(labels, ioc.Source)
	}

	return labels
}

// ConvertMultipleToBundle converts multiple IOCs into a single bundle
func (c *Converter) ConvertMultipleToBundle(iocs []IOCData, opts ConversionOptions) Bundle {
	objects := []any{}

	// Add organization identity once if needed
	if opts.IncludeIdentity && len(iocs) > 0 {
		now := FormatTimestamp(time.Now().UTC())
		identity := c.createIdentity(iocs[0], now)
		objects = append(objects, identity)

		// Set created by ref for all indicators
		for i := range iocs {
			indicator := c.createDomainIndicator(iocs[i], "indicator--"+uuid.New().String(), now)
			indicator.CreatedByRef = identity.ID
			objects = append(objects, indicator)
		}
	} else {
		// Just add indicators
		now := FormatTimestamp(time.Now().UTC())
		for _, ioc := range iocs {
			indicator := c.createDomainIndicator(ioc, "indicator--"+uuid.New().String(), now)
			objects = append(objects, indicator)
		}
	}

	bundleID := "bundle--" + uuid.New().String()
	return Bundle{
		Type:    "bundle",
		ID:      bundleID,
		Objects: objects,
	}
}

// GenerateStixID generates a new STIX indicator ID
func GenerateStixID() string {
	return "indicator--" + uuid.New().String()
}

// ValidateStixID checks if a STIX ID is valid format
func ValidateStixID(id string) bool {
	// Basic validation: should be "type--uuid"
	if len(id) < 10 {
		return false
	}
	// Should contain "--"
	if id[len(id)-37:len(id)-36] != "-" || id[len(id)-38:len(id)-37] != "-" {
		return false
	}
	return true
}

// buildSTIXPattern creates STIX 2.1 pattern based on IOC type
func (c *Converter) buildSTIXPattern(ioc IOCData) (string, string) {
	if ioc.Domain != "" {
		return fmt.Sprintf("[domain-name:value = '%s']", ioc.Domain), "domain"
	}
	if ioc.IP != "" {
		// Determine if IPv4 or IPv6
		if strings.Contains(ioc.IP, ":") {
			return fmt.Sprintf("[ipv6-addr:value = '%s']", ioc.IP), "ipv6"
		}
		return fmt.Sprintf("[ipv4-addr:value = '%s']", ioc.IP), "ipv4"
	}
	if ioc.URL != "" {
		return fmt.Sprintf("[url:value = '%s']", ioc.URL), "url"
	}
	if ioc.Hash != "" {
		hashType := strings.ToUpper(ioc.HashType)
		if hashType == "" {
			// Auto-detect
			switch len(ioc.Hash) {
			case 32:
				hashType = "MD5"
			case 40:
				hashType = "SHA-1"
			case 64:
				hashType = "SHA-256"
			default:
				hashType = "MD5"
			}
		} else {
			// Normalize hash type names
			switch strings.ToLower(ioc.HashType) {
			case "sha1":
				hashType = "SHA-1"
			case "sha256":
				hashType = "SHA-256"
			case "md5":
				hashType = "MD5"
			}
		}
		return fmt.Sprintf("[file:hashes.'%s' = '%s']", hashType, ioc.Hash), "file-hash"
	}
	if ioc.Email != "" {
		return fmt.Sprintf("[email-addr:value = '%s']", ioc.Email), "email"
	}
	// Fallback to domain if somehow nothing is set
	return "[domain-name:value = 'unknown']", "unknown"
}

// buildIndicatorName creates a human-readable name for the indicator
func (c *Converter) buildIndicatorName(ioc IOCData, iocType string) string {
	value := ""
	switch iocType {
	case "domain":
		value = ioc.Domain
	case "ipv4", "ipv6":
		value = ioc.IP
	case "url":
		value = ioc.URL
	case "file-hash":
		value = ioc.Hash
	case "email":
		value = ioc.Email
	default:
		value = "unknown"
	}

	return fmt.Sprintf("Malicious %s: %s", iocType, value)
}

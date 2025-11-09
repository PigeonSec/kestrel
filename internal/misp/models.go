package misp

// Attribute represents a MISP attribute
type Attribute struct {
	Type      string `json:"type"`
	Category  string `json:"category"`
	Value     string `json:"value"`
	ToIDs     bool   `json:"to_ids"`
	Comment   string `json:"comment,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// Event represents a MISP event
type Event struct {
	Info          string      `json:"info"`
	ThreatLevelID string      `json:"threat_level_id"`
	Analysis      string      `json:"analysis"`
	Distribution  string      `json:"distribution"`
	Attribute     []Attribute `json:"Attribute"`
}

// EventWrapper wraps a MISP event (standard MISP format)
type EventWrapper struct {
	Event Event `json:"Event"`
}

// ManifestEntry represents an entry in the MISP manifest
type ManifestEntry struct {
	UUID string `json:"uuid"`
}

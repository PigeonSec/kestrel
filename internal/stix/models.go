package stix

import "time"

// STIX 2.1 Core Object Types
// Spec: https://docs.oasis-open.org/cti/stix/v2.1/stix-v2.1.html

// Bundle represents a STIX 2.1 bundle containing objects
type Bundle struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Objects []any       `json:"objects"`
}

// Indicator represents a STIX 2.1 indicator object
type Indicator struct {
	Type             string   `json:"type"`
	SpecVersion      string   `json:"spec_version"`
	ID               string   `json:"id"`
	Created          string   `json:"created"`
	Modified         string   `json:"modified"`
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	IndicatorTypes   []string `json:"indicator_types,omitempty"`
	Pattern          string   `json:"pattern"`
	PatternType      string   `json:"pattern_type"`
	PatternVersion   string   `json:"pattern_version,omitempty"`
	ValidFrom        string   `json:"valid_from"`
	ValidUntil       string   `json:"valid_until,omitempty"`
	KillChainPhases  []KillChainPhase `json:"kill_chain_phases,omitempty"`
	Confidence       int      `json:"confidence,omitempty"`
	Labels           []string `json:"labels,omitempty"`
	ExternalReferences []ExternalReference `json:"external_references,omitempty"`
	ObjectMarkingRefs []string `json:"object_marking_refs,omitempty"`
	CreatedByRef     string   `json:"created_by_ref,omitempty"`
}

// Identity represents a STIX 2.1 identity object
type Identity struct {
	Type          string   `json:"type"`
	SpecVersion   string   `json:"spec_version"`
	ID            string   `json:"id"`
	Created       string   `json:"created"`
	Modified      string   `json:"modified"`
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	IdentityClass string   `json:"identity_class"`
	Sectors       []string `json:"sectors,omitempty"`
	ContactInformation string `json:"contact_information,omitempty"`
}

// MarkingDefinition represents TLP and other marking definitions
type MarkingDefinition struct {
	Type          string        `json:"type"`
	SpecVersion   string        `json:"spec_version"`
	ID            string        `json:"id"`
	Created       string        `json:"created"`
	DefinitionType string       `json:"definition_type"`
	Definition    TLPDefinition `json:"definition"`
	Name          string        `json:"name,omitempty"`
}

// TLPDefinition represents the TLP marking structure
type TLPDefinition struct {
	TLP string `json:"tlp"`
}

// Relationship represents a STIX 2.1 relationship object
type Relationship struct {
	Type          string   `json:"type"`
	SpecVersion   string   `json:"spec_version"`
	ID            string   `json:"id"`
	Created       string   `json:"created"`
	Modified      string   `json:"modified"`
	RelationshipType string `json:"relationship_type"`
	Description   string   `json:"description,omitempty"`
	SourceRef     string   `json:"source_ref"`
	TargetRef     string   `json:"target_ref"`
	StartTime     string   `json:"start_time,omitempty"`
	StopTime      string   `json:"stop_time,omitempty"`
}

// ObservedData represents observed cyber data
type ObservedData struct {
	Type          string         `json:"type"`
	SpecVersion   string         `json:"spec_version"`
	ID            string         `json:"id"`
	Created       string         `json:"created"`
	Modified      string         `json:"modified"`
	FirstObserved string         `json:"first_observed"`
	LastObserved  string         `json:"last_observed"`
	NumberObserved int           `json:"number_observed"`
	Objects       map[string]ObservableObject `json:"objects"`
}

// ObservableObject represents a cyber observable object
type ObservableObject struct {
	Type  string `json:"type"`
	Value string `json:"value,omitempty"`
	// Additional fields depending on type (domain-name, ipv4-addr, url, etc.)
}

// KillChainPhase represents a phase in a kill chain
type KillChainPhase struct {
	KillChainName string `json:"kill_chain_name"`
	PhaseName     string `json:"phase_name"`
}

// ExternalReference represents an external reference
type ExternalReference struct {
	SourceName  string `json:"source_name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	ExternalID  string `json:"external_id,omitempty"`
}

// Sighting represents when an indicator was observed
type Sighting struct {
	Type              string   `json:"type"`
	SpecVersion       string   `json:"spec_version"`
	ID                string   `json:"id"`
	Created           string   `json:"created"`
	Modified          string   `json:"modified"`
	FirstSeen         string   `json:"first_seen,omitempty"`
	LastSeen          string   `json:"last_seen,omitempty"`
	Count             int      `json:"count,omitempty"`
	SightingOfRef     string   `json:"sighting_of_ref"`
	ObservedDataRefs  []string `json:"observed_data_refs,omitempty"`
	WhereSightedRefs  []string `json:"where_sighted_refs,omitempty"`
	Summary           bool     `json:"summary,omitempty"`
	Description       string   `json:"description,omitempty"`
	Confidence        int      `json:"confidence,omitempty"`
	ObjectMarkingRefs []string `json:"object_marking_refs,omitempty"`
}

// ThreatActor represents a threat actor
type ThreatActor struct {
	Type              string   `json:"type"`
	SpecVersion       string   `json:"spec_version"`
	ID                string   `json:"id"`
	Created           string   `json:"created"`
	Modified          string   `json:"modified"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	ThreatActorTypes  []string `json:"threat_actor_types,omitempty"`
	Aliases           []string `json:"aliases,omitempty"`
	FirstSeen         string   `json:"first_seen,omitempty"`
	LastSeen          string   `json:"last_seen,omitempty"`
	Roles             []string `json:"roles,omitempty"`
	Goals             []string `json:"goals,omitempty"`
	Sophistication    string   `json:"sophistication,omitempty"`
	ResourceLevel     string   `json:"resource_level,omitempty"`
	PrimaryMotivation string   `json:"primary_motivation,omitempty"`
	SecondaryMotivations []string `json:"secondary_motivations,omitempty"`
	ObjectMarkingRefs []string `json:"object_marking_refs,omitempty"`
}

// Malware represents malware
type Malware struct {
	Type              string   `json:"type"`
	SpecVersion       string   `json:"spec_version"`
	ID                string   `json:"id"`
	Created           string   `json:"created"`
	Modified          string   `json:"modified"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	MalwareTypes      []string `json:"malware_types,omitempty"`
	IsFamily          bool     `json:"is_family"`
	Aliases           []string `json:"aliases,omitempty"`
	KillChainPhases   []KillChainPhase `json:"kill_chain_phases,omitempty"`
	FirstSeen         string   `json:"first_seen,omitempty"`
	LastSeen          string   `json:"last_seen,omitempty"`
	OperatingSystemRefs []string `json:"operating_system_refs,omitempty"`
	ArchitectureExecutionEnvs []string `json:"architecture_execution_envs,omitempty"`
	Capabilities      []string `json:"capabilities,omitempty"`
	SampleRefs        []string `json:"sample_refs,omitempty"`
	ObjectMarkingRefs []string `json:"object_marking_refs,omitempty"`
}

// AttackPattern represents an attack pattern (e.g., MITRE ATT&CK)
type AttackPattern struct {
	Type              string   `json:"type"`
	SpecVersion       string   `json:"spec_version"`
	ID                string   `json:"id"`
	Created           string   `json:"created"`
	Modified          string   `json:"modified"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	Aliases           []string `json:"aliases,omitempty"`
	KillChainPhases   []KillChainPhase `json:"kill_chain_phases,omitempty"`
	ExternalReferences []ExternalReference `json:"external_references,omitempty"`
	ObjectMarkingRefs []string `json:"object_marking_refs,omitempty"`
}

// Infrastructure represents adversary infrastructure
type Infrastructure struct {
	Type              string   `json:"type"`
	SpecVersion       string   `json:"spec_version"`
	ID                string   `json:"id"`
	Created           string   `json:"created"`
	Modified          string   `json:"modified"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	InfrastructureTypes []string `json:"infrastructure_types,omitempty"`
	Aliases           []string `json:"aliases,omitempty"`
	KillChainPhases   []KillChainPhase `json:"kill_chain_phases,omitempty"`
	FirstSeen         string   `json:"first_seen,omitempty"`
	LastSeen          string   `json:"last_seen,omitempty"`
	ObjectMarkingRefs []string `json:"object_marking_refs,omitempty"`
}

// Report represents a collection of threat intelligence
type Report struct {
	Type          string   `json:"type"`
	SpecVersion   string   `json:"spec_version"`
	ID            string   `json:"id"`
	Created       string   `json:"created"`
	Modified      string   `json:"modified"`
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	ReportTypes   []string `json:"report_types,omitempty"`
	Published     string   `json:"published"`
	ObjectRefs    []string `json:"object_refs"`
}

// Standard TLP Marking Definition IDs
const (
	TLPWHITE  = "marking-definition--613f2e26-407d-48c7-9eca-b8e91df99dc9"
	TLPGREEN  = "marking-definition--34098fce-860f-48ae-8e50-ebd3cc5e41da"
	TLPAMBER  = "marking-definition--f88d31f6-486f-44da-b317-01333bde0b82"
	TLPRED    = "marking-definition--5e57c739-391a-4eb3-b6be-7d15ca92d5ed"
)

// Helper function to format timestamps in STIX format (RFC3339)
func FormatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// Helper to get TLP marking ref from TLP string
func GetTLPMarkingRef(tlp string) string {
	switch tlp {
	case "TLP:WHITE", "WHITE":
		return TLPWHITE
	case "TLP:GREEN", "GREEN":
		return TLPGREEN
	case "TLP:AMBER", "AMBER":
		return TLPAMBER
	case "TLP:RED", "RED":
		return TLPRED
	default:
		return TLPAMBER // Default to AMBER
	}
}

// Envelope for TAXII responses
type Envelope struct {
	More    bool  `json:"more"`
	Next    string `json:"next,omitempty"`
	Objects []any `json:"objects"`
}

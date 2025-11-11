package misp

// Attribute represents a MISP attribute
type Attribute struct {
	ID               string   `json:"id,omitempty"`
	EventID          string   `json:"event_id,omitempty"`
	ObjectID         string   `json:"object_id,omitempty"`
	Type             string   `json:"type"`
	Category         string   `json:"category"`
	Value            string   `json:"value"`
	ToIDs            bool     `json:"to_ids"`
	UUID             string   `json:"uuid,omitempty"`
	Timestamp        string   `json:"timestamp,omitempty"`
	Distribution     string   `json:"distribution,omitempty"`
	SharingGroupID   string   `json:"sharing_group_id,omitempty"`
	Comment          string   `json:"comment,omitempty"`
	Deleted          bool     `json:"deleted,omitempty"`
	DisableCorrelation bool   `json:"disable_correlation,omitempty"`
	FirstSeen        string   `json:"first_seen,omitempty"`
	LastSeen         string   `json:"last_seen,omitempty"`
	Tag              []Tag    `json:"Tag,omitempty"`
}

// Object represents a MISP object (domain-ip, file, etc.)
type Object struct {
	ID               string      `json:"id,omitempty"`
	Name             string      `json:"name"`
	MetaCategory     string      `json:"meta-category"`
	Description      string      `json:"description,omitempty"`
	TemplateUUID     string      `json:"template_uuid,omitempty"`
	TemplateVersion  string      `json:"template_version,omitempty"`
	EventID          string      `json:"event_id,omitempty"`
	UUID             string      `json:"uuid,omitempty"`
	Timestamp        string      `json:"timestamp,omitempty"`
	Distribution     string      `json:"distribution,omitempty"`
	SharingGroupID   string      `json:"sharing_group_id,omitempty"`
	Comment          string      `json:"comment,omitempty"`
	Deleted          bool        `json:"deleted,omitempty"`
	FirstSeen        string      `json:"first_seen,omitempty"`
	LastSeen         string      `json:"last_seen,omitempty"`
	Attribute        []Attribute `json:"Attribute"`
	ObjectReference  []Reference `json:"ObjectReference,omitempty"`
}

// Reference represents a relationship between objects
type Reference struct {
	ID               string `json:"id,omitempty"`
	UUID             string `json:"uuid,omitempty"`
	Timestamp        string `json:"timestamp,omitempty"`
	ObjectID         string `json:"object_id"`
	ReferencedUUID   string `json:"referenced_uuid"`
	ReferencedID     string `json:"referenced_id"`
	ReferencedType   string `json:"referenced_type"`
	RelationshipType string `json:"relationship_type"`
	Comment          string `json:"comment,omitempty"`
	Deleted          bool   `json:"deleted,omitempty"`
}

// Tag represents a MISP tag
type Tag struct {
	ID               string `json:"id,omitempty"`
	Name             string `json:"name"`
	Colour           string `json:"colour,omitempty"`
	Exportable       bool   `json:"exportable,omitempty"`
	OrgID            string `json:"org_id,omitempty"`
	UserID           string `json:"user_id,omitempty"`
	HideTag          bool   `json:"hide_tag,omitempty"`
	NumericalValue   string `json:"numerical_value,omitempty"`
}

// Galaxy represents threat actor/malware families
type Galaxy struct {
	ID          string        `json:"id,omitempty"`
	UUID        string        `json:"uuid,omitempty"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Description string        `json:"description,omitempty"`
	Version     string        `json:"version,omitempty"`
	Icon        string        `json:"icon,omitempty"`
	Namespace   string        `json:"namespace,omitempty"`
	GalaxyCluster []GalaxyCluster `json:"GalaxyCluster,omitempty"`
}

// GalaxyCluster represents specific threat actors/campaigns
type GalaxyCluster struct {
	ID          string            `json:"id,omitempty"`
	UUID        string            `json:"uuid,omitempty"`
	CollectionUUID string         `json:"collection_uuid,omitempty"`
	Type        string            `json:"type"`
	Value       string            `json:"value"`
	Tag         string            `json:"tag_name,omitempty"`
	Description string            `json:"description,omitempty"`
	Source      string            `json:"source,omitempty"`
	Authors     []string          `json:"authors,omitempty"`
	Version     string            `json:"version,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// Sighting represents when an indicator was observed
type Sighting struct {
	ID        string `json:"id,omitempty"`
	AttributeID string `json:"attribute_id,omitempty"`
	EventID   string `json:"event_id,omitempty"`
	OrgID     string `json:"org_id,omitempty"`
	DateSighting string `json:"date_sighting"`
	UUID      string `json:"uuid,omitempty"`
	Source    string `json:"source,omitempty"`
	Type      string `json:"type,omitempty"` // 0=sighting, 1=false-positive, 2=expiration
}

// Event represents a MISP event
type Event struct {
	ID            string      `json:"id,omitempty"`
	OrgID         string      `json:"org_id,omitempty"`
	OrgcID        string      `json:"orgc_id,omitempty"`
	Date          string      `json:"date"`
	ThreatLevelID string      `json:"threat_level_id"`
	Info          string      `json:"info"`
	Published     bool        `json:"published,omitempty"`
	UUID          string      `json:"uuid,omitempty"`
	AttributeCount string     `json:"attribute_count,omitempty"`
	Analysis      string      `json:"analysis"`
	Timestamp     string      `json:"timestamp"`
	Distribution  string      `json:"distribution"`
	ProposalEmailLock bool    `json:"proposal_email_lock,omitempty"`
	Locked        bool        `json:"locked,omitempty"`
	PublishTimestamp string   `json:"publish_timestamp,omitempty"`
	SharingGroupID string     `json:"sharing_group_id,omitempty"`
	DisableCorrelation bool   `json:"disable_correlation,omitempty"`
	ExtendsUUID   string      `json:"extends_uuid,omitempty"`
	Org           *Org        `json:"Org,omitempty"`
	Orgc          *Org        `json:"Orgc,omitempty"`
	Attribute     []Attribute `json:"Attribute"`
	Object        []Object    `json:"Object,omitempty"`
	Galaxy        []Galaxy    `json:"Galaxy,omitempty"`
	Tag           []Tag       `json:"Tag,omitempty"`
	RelatedEvent  []RelatedEvent `json:"RelatedEvent,omitempty"`
}

// Org represents a MISP organization
type Org struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	UUID  string `json:"uuid,omitempty"`
	Local bool   `json:"local,omitempty"`
}

// RelatedEvent represents related MISP events
type RelatedEvent struct {
	Event Event `json:"Event"`
}

// EventWrapper wraps a MISP event (standard MISP format)
type EventWrapper struct {
	Event Event `json:"Event"`
}

// ManifestEntry represents an entry in the MISP manifest
type ManifestEntry struct {
	UUID      string `json:"uuid"`
	Info      string `json:"info,omitempty"`
	Date      string `json:"date,omitempty"`
	ThreatLevelID string `json:"threat_level_id,omitempty"`
	Analysis  string `json:"analysis,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

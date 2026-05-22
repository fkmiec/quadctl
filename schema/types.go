package schema

import (
	"text/template"
)

// OptionValue represents a valid value for an option
type OptionValue struct {
	Value       string `json:"value"`
	Description string `json:"description"`
	Info        string `json:"info,omitempty"`
	Validator   string `json:"validator"`
}

// SchemaOption represents a single option in the schema
type SchemaOption struct {
	QuadletKey            string             `json:"quadlet-key"`
	PodmanKey             string             `json:"podman-key"`
	Description           string             `json:"description"`
	QuadletTemplate       string             `json:"quadlet-template"`
	PodmanTemplate        string             `json:"podman-template"`
	AllowMultiple         bool               `json:"allow-multiple"`
	Values                []OptionValue      `json:"values"`
	QuadletTemplateParsed *template.Template `json:"-"`
	PodmanTemplateParsed  *template.Template `json:"-"`
}

// SchemaType represents the schema for a unit type
type SchemaType struct {
	Type    string         `json:"type"`
	Options []SchemaOption `json:"options"`
}

// Schema represents the complete schema
type Schema []SchemaType

// OptionMetadata holds extracted metadata for an option
type OptionMetadata struct {
	QuadletKey     string
	QuadletFormat  string // Format string with placeholders
	PodmanKey      string
	PodmanFormat   string // Format string with placeholders
	Description    string
	AllowMultiple  bool
	KnownValues    []OptionValue
	ValidatorRegex string
	ValueType      string // e.g., "ipv4", "integer", "duration", "path", "capability"
}

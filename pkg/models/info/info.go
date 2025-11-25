package info

import "encoding/json"

// VersionInfo represents version and build information for an OSDU service
type VersionInfo struct {
	GroupID                string                  `json:"groupId,omitempty"`
	ArtifactID             string                  `json:"artifactId,omitempty"`
	Version                string                  `json:"version,omitempty"`
	BuildTime              string                  `json:"buildTime,omitempty"`
	Branch                 string                  `json:"branch,omitempty"`
	CommitID               string                  `json:"commitId,omitempty"`
	CommitMessage          string                  `json:"commitMessage,omitempty"`
	ConnectedOuterServices []ConnectedOuterService `json:"connectedOuterServices,omitempty"`
	FeatureFlagStates      []FeatureFlagState      `json:"featureFlagStates,omitempty"`
}

// ConnectedOuterService represents an outer service connected to the OSDU service
// Contains service-specific values for all outer services connected to OSDU service
type ConnectedOuterService struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// FeatureFlagState represents the state of a feature flag
type FeatureFlagState struct {
	Name      string `json:"name,omitempty"`
	Enabled   bool   `json:"enabled"`
	Partition string `json:"partition,omitempty"`
	Source    string `json:"source,omitempty"`
}

// UnstructuredVersionInfo is an alternative representation with flexible fields
// Use this when you need to handle additional fields not in the standard model
type UnstructuredVersionInfo struct {
	GroupID                string                 `json:"groupId,omitempty"`
	ArtifactID             string                 `json:"artifactId,omitempty"`
	Version                string                 `json:"version,omitempty"`
	BuildTime              string                 `json:"buildTime,omitempty"`
	Branch                 string                 `json:"branch,omitempty"`
	CommitID               string                 `json:"commitId,omitempty"`
	CommitMessage          string                 `json:"commitMessage,omitempty"`
	ConnectedOuterServices json.RawMessage        `json:"connectedOuterServices,omitempty"`
	FeatureFlagStates      json.RawMessage        `json:"featureFlagStates,omitempty"`
	AdditionalFields       map[string]interface{} `json:"-"`
}

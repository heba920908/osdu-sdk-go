package v2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/heba920908/osdu-sdk-go/pkg/models/dataset"
	"github.com/heba920908/osdu-sdk-go/pkg/models/info"
)

// DatasetService defines the interface for dataset operations
type DatasetService interface {
	// GetInfo retrieves version information without authorization
	GetInfo() (*info.VersionInfo, error)

	// GetStorageInstructions retrieves storage instructions for uploading a dataset
	GetStorageInstructions(kindSubType string) (*dataset.DatasetStorageInstructionsResponse, error)

	// GetRetrievalInstructions retrieves retrieval instructions for downloading datasets
	GetRetrievalInstructions(datasetRegistryIds []string) (*dataset.DatasetRetrievalInstructionsResponse, error)
}

// datasetService implements the DatasetService interface
type datasetService struct {
	client *Client
}

// GetInfo retrieves version information
// This endpoint does not require authorization
func (s *datasetService) GetInfo() (*info.VersionInfo, error) {
	url := fmt.Sprintf("%s/info", s.client.osduSettings.DatasetUrl)
	s.client.logger.Debug("Getting dataset info", "url", url)

	resp, err := s.client.doRequest(http.MethodGet, url, nil, false, false)
	if err != nil {
		s.client.logger.Error("Failed to get dataset info", "error", err)
		return nil, fmt.Errorf("failed to get dataset info: %w", err)
	}

	body, err := readResponse(resp)
	if err != nil {
		s.client.logger.Error("Failed to read dataset info response", "error", err, "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to read dataset info response: %w", err)
	}

	var versionInfo info.VersionInfo
	if err := json.Unmarshal(body, &versionInfo); err != nil {
		s.client.logger.Error("Failed to unmarshal version info", "error", err)
		return nil, fmt.Errorf("failed to unmarshal version info: %w", err)
	}

	s.client.logger.Info("Retrieved dataset info",
		"version", versionInfo.Version,
		"artifactId", versionInfo.ArtifactID)
	return &versionInfo, nil
}

// GetStorageInstructions retrieves storage instructions for uploading a dataset
// This endpoint requires authorization
func (s *datasetService) GetStorageInstructions(kindSubType string) (*dataset.DatasetStorageInstructionsResponse, error) {
	url := fmt.Sprintf("%s/getStorageInstructions?kindSubType=%s", s.client.osduSettings.DatasetUrl, kindSubType)
	s.client.logger.Debug("Getting storage instructions", "url", url, "kindSubType", kindSubType)

	resp, err := s.client.doRequest(http.MethodPost, url, nil, true, true)
	if err != nil {
		s.client.logger.Error("Failed to get storage instructions", "error", err)
		return nil, fmt.Errorf("failed to get storage instructions: %w", err)
	}

	body, err := readResponse(resp)
	if err != nil {
		s.client.logger.Error("Failed to read storage instructions response", "error", err, "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to read storage instructions response: %w", err)
	}

	var storageInstructions dataset.DatasetStorageInstructionsResponse
	if err := json.Unmarshal(body, &storageInstructions); err != nil {
		s.client.logger.Error("Failed to unmarshal storage instructions", "error", err)
		return nil, fmt.Errorf("failed to unmarshal storage instructions: %w", err)
	}

	s.client.logger.Info("Retrieved storage instructions",
		"providerKey", storageInstructions.ProviderKey,
		"signedUrl", storageInstructions.StorageLocation.SignedUrl)
	return &storageInstructions, nil
}

// GetRetrievalInstructions retrieves retrieval instructions for downloading datasets
// This endpoint requires authorization
func (s *datasetService) GetRetrievalInstructions(datasetRegistryIds []string) (*dataset.DatasetRetrievalInstructionsResponse, error) {
	url := fmt.Sprintf("%s/getRetrievalInstructions", s.client.osduSettings.DatasetUrl)
	s.client.logger.Debug("Getting retrieval instructions", "url", url, "datasetRegistryIds", datasetRegistryIds)

	reqBody, err := json.Marshal(dataset.DatasetRetrievalRequest{
		DatasetRegistryIds: datasetRegistryIds,
	})
	if err != nil {
		s.client.logger.Error("Failed to marshal retrieval request", "error", err)
		return nil, fmt.Errorf("failed to marshal retrieval request: %w", err)
	}

	resp, err := s.client.doRequest(http.MethodPost, url, reqBody, true, true)
	if err != nil {
		s.client.logger.Error("Failed to get retrieval instructions", "error", err)
		return nil, fmt.Errorf("failed to get retrieval instructions: %w", err)
	}

	body, err := readResponse(resp)
	if err != nil {
		s.client.logger.Error("Failed to read retrieval instructions response", "error", err, "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to read retrieval instructions response: %w", err)
	}

	var retrievalInstructions dataset.DatasetRetrievalInstructionsResponse
	if err := json.Unmarshal(body, &retrievalInstructions); err != nil {
		s.client.logger.Error("Failed to unmarshal retrieval instructions", "error", err)
		return nil, fmt.Errorf("failed to unmarshal retrieval instructions: %w", err)
	}

	s.client.logger.Info("Retrieved retrieval instructions",
		"datasetCount", len(retrievalInstructions.Datasets))
	return &retrievalInstructions, nil
}


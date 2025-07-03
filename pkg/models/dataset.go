package models

// OSDUDatasetResponse represents the structure of the OSDU dataset output response
type DatasetStorageInstructionsResponse struct {
	ProviderKey     string                 `json:"providerKey"`
	StorageLocation DatasetStorageLocation `json:"storageLocation"`
}

// StorageLocation represents the storage location details within the OSDU dataset response
type DatasetStorageLocation struct {
	SignedUrl  string `json:"signedUrl"`
	FileSource string `json:"fileSource,omitempty"`
	CreatedBy  string `json:"createdBy"`
	ExpiryTime string `json:"expiryTime"`
}

type DatasetRetrievalRequest struct {
	DatasetRegistryIds []string `json:"datasetRegistryIds"`
}

type DatasetRetrievalInstructionsResponse struct {
	Datasets []Dataset `json:"datasets"`
}

// Dataset represents individual dataset details
type Dataset struct {
	DatasetRegistryId   string                 `json:"datasetRegistryId"`
	RetrievalProperties DatasetStorageLocation `json:"retrievalProperties"`
	ProviderKey         string                 `json:"providerKey"`
}

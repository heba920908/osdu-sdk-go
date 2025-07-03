package models

// OSDUDatasetResponse represents the structure of the OSDU dataset output response
type DatasetResponse struct {
	ProviderKey     string                 `json:"providerKey"`
	StorageLocation DatasetStorageLocation `json:"storageLocation"`
}

// StorageLocation represents the storage location details within the OSDU dataset response
type DatasetStorageLocation struct {
	SignedUrl  string `json:"signedUrl"`
	FileSource string `json:"fileSource"`
	CreatedBy  string `json:"createdBy"`
	ExpiryTime string `json:"expiryTime"`
}

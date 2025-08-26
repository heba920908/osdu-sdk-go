package models

type FileGetSignedUrlResponse struct {
	FileID               string `json:"FileID"`
	FileResponseLocation `json:"Location"`
}

type FileResponseLocation struct {
	SignedUrl  string `json:"SignedUrl"`
	FileSource string `json:"FileSource,omitempty"`
}

type FileGetDownloadUrlResponse struct {
	SignedUrl string `json:"SignedUrl"`
}

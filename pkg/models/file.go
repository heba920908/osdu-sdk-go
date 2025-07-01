package models

type FileGetSignedUrlResponse struct {
	FileID               string `json:"FileID"`
	FileResponseLocation `json:"Location"`
}

type FileResponseLocation struct {
	SignedURL  string `json:"SignedURL"`
	FileSource string `json:"FileSource"`
}

type FileGetDownloadUrlResponse struct {
	SignedUrl string `json:"SignedUrl"`
}

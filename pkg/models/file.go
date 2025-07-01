package models

type FileGetSignedUrlResponse struct {
	FileID   string `json:"FileID"`
	Location struct {
		SignedURL  string `json:"SignedURL"`
		FileSource string `json:"FileSource"`
	} `json:"Location"`
}

type FileGetDownloadUrlResponse struct {
	SignedUrl string `json:"SignedUrl"`
}

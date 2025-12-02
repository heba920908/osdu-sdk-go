package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	notificationService = "de-notification-service"
	expireDuration      = 30000 // 30 seconds in milliseconds
)

// HmacData represents the data structure embedded in HMAC signature
// Matches the Java structure from SignatureService
type HmacData struct {
	ExpireMillisecond string `json:"expireMillisecond"`
	HashMechanism     string `json:"hashMechanism"`
	EndpointUrl       string `json:"endpointUrl"`
	Nonce             string `json:"nonce"`
}

// SignatureService provides HMAC signature generation and verification
// This is a Go implementation of Java's SignatureService
type SignatureService struct{}

// NewSignatureService creates a new signature service
func NewSignatureService() *SignatureService {
	return &SignatureService{}
}

// VerifyHmacSignature verifies the HMAC signature
// hmac format: "base64(data).hexSignature"
// This implements the logic from SignatureService.java:verifyHmacSignature
func (s *SignatureService) VerifyHmacSignature(hmacStr, secret string) error {
	if hmacStr == "" {
		return fmt.Errorf("HMAC signature should not be null or empty")
	}
	if secret == "" {
		return fmt.Errorf("secret should not be null or empty")
	}

	// Split the HMAC into data and signature parts
	tokens := strings.Split(hmacStr, ".")
	if len(tokens) != 2 {
		return fmt.Errorf("invalid signature format")
	}

	// Decode the base64 data part
	dataBytes, err := base64.StdEncoding.DecodeString(tokens[0])
	if err != nil {
		return fmt.Errorf("failed to decode signature data: %w", err)
	}

	requestSignature := tokens[1]

	// Parse the JSON data
	var hmacData HmacData
	if err := json.Unmarshal(dataBytes, &hmacData); err != nil {
		return fmt.Errorf("failed to parse HMAC data: %w", err)
	}

	// Validate required fields
	if hmacData.EndpointUrl == "" || hmacData.Nonce == "" || hmacData.ExpireMillisecond == "" {
		return fmt.Errorf("missing url or nonce or expire time in the signature")
	}

	// Generate the expected signature using the same parameters
	newSignature, err := s.GetSignedSignature(hmacData.EndpointUrl, secret, hmacData.ExpireMillisecond, hmacData.Nonce)
	if err != nil {
		return fmt.Errorf("failed to generate signature: %w", err)
	}

	// Compare signatures (case-insensitive)
	if !strings.EqualFold(requestSignature, newSignature) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// GetSignedSignature generates a signature for the given parameters
// This is used for verification - regenerates signature with provided expireTime and nonce
func (s *SignatureService) GetSignedSignature(url, secret, expireTime, nonce string) (string, error) {
	if url == "" || secret == "" || expireTime == "" {
		return "", fmt.Errorf("error generating the signature")
	}

	// Parse and validate expireTime
	expiry, err := strconv.ParseInt(expireTime, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid expire time: %w", err)
	}

	// Check if signature is expired
	if time.Now().UnixMilli() > expiry {
		return "", fmt.Errorf("signature is expired")
	}

	// Calculate timestamp from expiry
	timeStamp := strconv.FormatInt(expiry-expireDuration, 10)

	// Build the data JSON
	data := fmt.Sprintf(`{"expireMillisecond": "%s","hashMechanism": "hmacSHA256","endpointUrl": "%s","nonce": "%s"}`,
		expireTime, url, nonce)

	// Compute the signature
	signature, err := s.computeSignature(secret, nonce, timeStamp, data)
	if err != nil {
		return "", fmt.Errorf("error generating the signature: %w", err)
	}

	// Return hex-encoded signature
	return hex.EncodeToString(signature), nil
}

// computeSignature computes the multi-layer HMAC signature
// This implements the logic from SignatureService.java:getSignature
func (s *SignatureService) computeSignature(secret, nonce, timeStamp, data string) ([]byte, error) {
	// Step 1: Parse secret and nonce from hex
	secretBytes, err := hex.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode secret: %w", err)
	}

	nonceBytes, err := hex.DecodeString(nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	// Step 2: Encrypt nonce with secret
	encryptedNonce := computeHmacSha256Bytes(nonceBytes, secretBytes)

	// Step 3: Encrypt timestamp with encrypted nonce
	encryptedTimestamp := computeHmacSha256String(timeStamp, encryptedNonce)

	// Step 4: Create signed key with notification service constant
	signedKey := computeHmacSha256String(notificationService, encryptedTimestamp)

	// Step 5: Sign the data with the signed key
	signature := computeHmacSha256String(data, signedKey)

	return signature, nil
}

// computeHmacSha256String computes HMAC-SHA256 for string data
func computeHmacSha256String(data string, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// computeHmacSha256Bytes computes HMAC-SHA256 for byte data
func computeHmacSha256Bytes(data, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

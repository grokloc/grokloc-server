// Package jwt provides token-related functionality
package jwt

import (
	"crypto/sha256"
	"fmt"
)

func EncodeTokenRequest(id, apiSecret string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(id+apiSecret)))
}

func VerifyTokenRequest(request, id, apiSecret string) bool {
	return EncodeTokenRequest(id, apiSecret) == request
}

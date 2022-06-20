// Package jwt provides token-related functionality
package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	jwt_go "github.com/dgrijalva/jwt-go"
	"github.com/grokloc/grokloc-server/pkg/security"
)

// JWT related constants
const (
	Authorization = "Authorization"
	TokenType     = "Bearer"
	Expiration    = 86400
)

// Claims are the JWT claims for the app
type Claims struct {
	Scope string `json:"scope"`
	Org   string `json:"org"`
	jwt_go.StandardClaims
}

// EncodeTokenRequest builds the token request from the user id and api secret
func EncodeTokenRequest(userID, userApiSecret string) string {
	return security.EncodedSHA256(userID + userApiSecret)
}

// Verify TokenRequest vaidates the request against the user id and api secret
func VerifyTokenRequest(request, userID, userApiSecret string) bool {
	return EncodeTokenRequest(userID, userApiSecret) == request
}

// New returns a new Claims instance
func New(userID, userEmailDigest, orgID string) (*Claims, error) {
	now := time.Now().Unix()
	claims := &Claims{
		"app",
		orgID,
		jwt_go.StandardClaims{
			Audience:  userEmailDigest,
			ExpiresAt: now + int64(Expiration),
			Id:        userID,
			Issuer:    "GrokLOC.com",
			IssuedAt:  now,
		}}
	return claims, nil
}

// ToHeaderVal prepends the JWTTokenType
func ToHeaderVal(token string) string {
	return fmt.Sprintf("%s %s", TokenType, token)
}

// FromHeaderVal will remove the JWTTokenType if it prepends the string s,
// but is also safe to use if s is just the token
func FromHeaderVal(s string) string {
	return strings.TrimPrefix(s, fmt.Sprintf("%s ", TokenType))
}

// Decode returns the claims from a signed string jwt
func Decode(id, token string, signingKey []byte) (*Claims, error) {
	f := func(token *jwt_go.Token) (interface{}, error) {
		return []byte(id + string(signingKey)), nil
	}
	parsed, err := jwt_go.ParseWithClaims(token, &Claims{}, f)
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if ok {
		return claims, nil
	}
	return nil, errors.New("token claims")
}

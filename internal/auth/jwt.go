package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AdminClaims defines the structural payload hidden inside the encrypted token string
type AdminClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed, cryptographically secure 1-hour token string
func GenerateToken(username string, secretKey string) (string, error) {
	// 1. Pack your claims with your identity metadata and set the 1-hour expiration window
	claims := AdminClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "mrunalini-art-portfolio",
		},
	}

	// 2. Specify the cryptographic signing method (HMAC SHA256 is the standard web norm)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. Digitally sign the token using your private environmental secret key string
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateToken parses and cryptographically verifies the validity of an incoming token string
func ValidateToken(tokenString string, secretKey string) (*AdminClaims, error) {
	// 1. Parse the incoming string using our structural claims blueprint layout mapping rules
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure that the token's signing algorithm matches what we expect (HMAC)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected cryptographic signing method algorithm utilized")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 2. Check if the token claims block unpacking logic executed successfully and token signature is active
	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid or expired token session token string tracking details")
}

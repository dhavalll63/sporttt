// pkg/token/token.go
package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5" // Using v5
)

// Claims defines the structure of the JWT claims your application uses.
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role,omitempty"` // Role can be included in JWT for quick checks, but DB is source of truth
	jwt.RegisteredClaims
}

// ValidateJWT parses, validates, and returns claims from a JWT string.
func ValidateJWT(tokenString string, secretKey string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("token string is empty")
	}
	if secretKey == "" {
		return nil, errors.New("jwt secret key is empty")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, errors.New("token is not yet valid")
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, errors.New("token signature is invalid")
		}
		return nil, fmt.Errorf("could not parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	if claims.UserID == 0 { // Basic validation for your custom claim
		return nil, errors.New("user_id claim is missing or zero")
	}

	// Check 'exp' claim for expiry, which ParseWithClaims should do, but an explicit check is fine.
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired (checked manually)")
	}

	return claims, nil
}

// GenerateJWT (Example you might already have or a similar one)
func GenerateJWT(userID uint, userRole string, secretKey string, expiryMinutes int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)
	claims := &Claims{
		UserID: userID,
		Role:   userRole, // Optionally include role
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app-name", // Optional
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

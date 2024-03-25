package models

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	RefreshTokenExpiry = 24 * 7 * time.Hour // 7 days
	AccessTokenExpiry  = 15 * time.Minute   // 15 minute
	// RefreshTokenExpiry = 120 * time.Second // for test
	// AccessTokenExpiry  = 3 * time.Second   // for test
)

type CustomClaim struct {
	jwt.StandardClaims
	UserID   int
	Username string
	Role string
}

type Users struct {
	UserID         int    		`json:"user_id"`
	Username       string 		`json:"username"`
	Email          string 		`json:"email"`
	HashedPassword string 		`json:"hashed_password"`
	Phone          string 		`json:"phone"`
	Role           string 		`json:"role"`
	CreatedAt      time.Time 	`json:"created_at"`
}

type RefreshToken struct {
    TokenID       string    `json:"tokenId"`       // Assuming TokenID is a string (UUID format)
    FamilyID      string    `json:"familyId"`      // FamilyID to link tokens in the same refresh family
    IssuedAt      time.Time `json:"issuedAt"`      // IssuedAt to store the timestamp when the token was issued
    ExpiresAt     time.Time `json:"expiresAt"`     // ExpiresAt to store the timestamp when the token will expire
    RevokesAt     time.Time `json:"revokedAt"`       // Revoked indicates whether the token has been revoked
}

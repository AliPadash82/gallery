package MyJwt

import (
	"crypto/rsa"
	"fmt"
	"main/database"
	"main/models"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	privateKeyPath     = "keys/app.rsa"
	publicKeyPath      = "keys/app.rsa.pub"
	CSRF_TOKEN_NAME    = "X-CSRF-Token"
	REFRESH_TOKEN_NAME = "RefreshToken"
	ACCESS_TOKEN_NAME  = "AuthToken"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

// InitJwt initializes the JWT keys used for signing and verification.
func InitJwt() (err error) {
	pemBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return
	}
	pubKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return
	}
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	return
}

func NullifyTokens(w http.ResponseWriter) {
	RemoveCSRFCookie(w)
	RemoveAccessTokenCookie(w)
	RemoveRefreshTokenCookie(w)
}

func IssueNewAccessToken(w http.ResponseWriter, r *http.Request, userID int, username, role string) (err error) {
	// Validate the refresh token and extract claims
	claims, err := ValidateRefreshTokenAndGetClaims(r)
	if err != nil {
		return
	}

	err = database.RevokeTokenWithTokenID(claims.Id)
	if err != nil {
		return
	}

	// Generate a new refresh token
	familyID, err := database.GetFamilyIDFromTokenID(claims.Id)
	if err != nil {
		return
	}

	var refreshTokenString string
	err = executeWithRetry(func() (innerErr error) {
		fmt.Println("renewing refresh token")
		refreshTokenString, innerErr = GenerateRefreshToken(familyID)
		if innerErr != nil {
			return
		}
		SetRefreshTokenCookie(w, refreshTokenString)
		return
	}, 5, time.Second)

	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return err
	}

	// Generate a new access token
	accessTokenString, err := GenerateAccessToken(userID, username, role)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	// Set the new access token in an HTTP-only cookie
	SetAccessTokenCookie(w, accessTokenString)
	return
}

func executeWithRetry(operation func() error, maxAttempts int, retryDelay time.Duration) error {
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = operation()
		if err == nil {
			return nil // Success, no need to retry
		}
		time.Sleep(retryDelay)
	}
	return err
}

func IssueTokens(w http.ResponseWriter, newUser models.Users) {
	accessToken, err := GenerateAccessToken(newUser.UserID, newUser.Username, newUser.Role)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}
	refreshToken, err := GenerateRefreshToken("")
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	SetAccessTokenCookie(w, accessToken)
	SetRefreshTokenCookie(w, refreshToken)
}

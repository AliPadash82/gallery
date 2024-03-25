package MyJwt

import (
	"fmt"
	"main/models"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateAuthToken generates a new access token for the given user ID
func GenerateAccessToken(userId int, username string, role string) (string, error) {
	// Create the access token claims with the user ID and the expiration time
	accessTokenClaims := &models.CustomClaim{
		UserID:     userId,
		Username:   username,
		Role:       role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(models.AccessTokenExpiry).Unix(),
		},
	}

	// Create a new JWT token with the claims and the signing method
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessTokenClaims)

	// Generate the signed string using the private key
	accessTokenString, err := accessToken.SignedString(signKey)
	if err != nil {
		return "", err
	}

	return accessTokenString, nil
}

func SetAccessTokenCookie(w http.ResponseWriter, accessTokenString string) {
	cookie := &http.Cookie{
		Name:     ACCESS_TOKEN_NAME,
		Value:    accessTokenString,
		HttpOnly: true,
		Path:     "/", // Ensure the path matches the one used when setting the cookie
		MaxAge:   int(models.RefreshTokenExpiry.Seconds()),
	}
	http.SetCookie(w, cookie)
}

func GetAccessTokenFromRequest(r *http.Request) (*http.Cookie, error) {
	fmt.Print(ACCESS_TOKEN_NAME)
	cookie, err := r.Cookie(ACCESS_TOKEN_NAME)
	if err != nil {
		return nil, err
	}
	return cookie, nil
}

func RemoveAccessTokenCookie(w http.ResponseWriter) {
	// Create a cookie with the same name as the access token cookie,
	// but set its expiration to the past and value to an empty string.
	cookie := &http.Cookie{
		Name:     ACCESS_TOKEN_NAME,
		Value:    "",
		Expires:  time.Unix(0, 0), // Set to a time in the past
		Path:     "/",             // Ensure the path matches the one used when setting the cookie
		HttpOnly: true,            // Should match the original cookie's attributes
	}
	http.SetCookie(w, cookie)
}

func ValidateAccessToken(r *http.Request) (*jwt.Token, error) {
	accessTokenCookie, err := GetAccessTokenFromRequest(r)
	if err != nil {
		return nil, err
	}
	// Parse the token using the public key
	return jwt.ParseWithClaims(accessTokenCookie.Value, &models.CustomClaim{}, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the public key
		return verifyKey, nil
	})
}

func GetAccessClaims(token *jwt.Token) (*models.CustomClaim, error) {
	claims, ok := token.Claims.(*models.CustomClaim)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func ValidateAccessTokenAndGetClaims(r *http.Request) (*models.CustomClaim, error) {
	token, err := ValidateAccessToken(r)
	if err != nil {
		return nil, err
	}
	claims, err := GetAccessClaims(token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

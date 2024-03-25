package MyJwt

import (
	"database/sql"
	"errors"
	"fmt"
	"main/database"
	"main/models"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func SetRefreshTokenCookie(w http.ResponseWriter, refreshTokenString string) {
	cookie := &http.Cookie{
		Name:     REFRESH_TOKEN_NAME,
		Value:    refreshTokenString,
		HttpOnly: true,
		Path:     "/", // Ensure the path matches the one used when setting the cookie
		MaxAge:   int(models.RefreshTokenExpiry.Seconds()),
	}

	http.SetCookie(w, cookie)
}

func GenerateUniqueRefreshTokenID() (id string, err error) {
    attempts := 3
    for id == "" && attempts > 0 {
        id, err = GenerateRandomString()
        if err != nil {
            // Error generating the random string; continue to the next attempt
            id = ""
			continue
        }
        _, err = database.GetRefreshToken(id)
        if err != nil {
            if err == sql.ErrNoRows {
                // This ID does not exist in the database, so it's unique and acceptable
                return id, nil
            }
            // An unexpected error occurred while trying to fetch the refresh token
			attempts--
			id = ""
            continue
        }
		return "", err
    }
    return "", errors.New("could not generate unique refresh token ID")
}


// GenerateRefreshToken generates a new refresh token for the given user ID
func GenerateRefreshToken(familyID string) (string, error) {
	id, err := GenerateUniqueRefreshTokenID()
	if err != nil {
		return "", err
	}
	if familyID == "" {
		familyID = id
	}
	refreshTokenClaims := &models.CustomClaim{
		StandardClaims: jwt.StandardClaims{
			Id:        id,
			ExpiresAt: time.Now().Add(models.RefreshTokenExpiry).Unix(),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(signKey)
	if err != nil {
		return "", err
	}
	err = database.StoreRefreshToken(
		models.RefreshToken{
		TokenID: id,
		FamilyID: familyID,
		IssuedAt: time.Now(),
		ExpiresAt: time.Now().Add(models.RefreshTokenExpiry),
		RevokesAt: time.Now().Add(models.RefreshTokenExpiry),
	})
	if err != nil {
		return "", err
	}
	return refreshTokenString, nil
}

func GetRefreshTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(REFRESH_TOKEN_NAME)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func RemoveRefreshTokenCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     REFRESH_TOKEN_NAME,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
}

// ValidateTokenAndGetClaims validates the given token string
func ValidateRefreshTokenAndGetClaims(r *http.Request) (*models.CustomClaim, error) {
	refreshTokenValue, err := GetRefreshTokenFromRequest(r)
	if err != nil {
		return nil, err
	}
	// Parse the token using the public key
	token, err := jwt.ParseWithClaims(refreshTokenValue, &models.CustomClaim{}, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the public key
		return verifyKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	claims, ok := token.Claims.(*models.CustomClaim)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	revoked, err := database.VerifyRefreshToken(claims.Id)
	if err != nil {
		return nil, err
	}
	if revoked {
		database.RevokeTokenFamilyWithTokenID(claims.Id)
		return nil, errors.New("refresh token has been revoked")
	}

	return claims, nil
}

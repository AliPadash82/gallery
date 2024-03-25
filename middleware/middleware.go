package middleware

import (
	"fmt"
	myjwt "main/MyJwt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func AuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate the token from the AuthToken cookie
		accessToken, err := myjwt.ValidateAccessToken(r)
		if err != nil {
			fmt.Println(err)
			// Check if the error is because the cookie does not exist
			if err == http.ErrNoCookie {
				fmt.Println("AuthToken cookie does not exist")
				myjwt.NullifyTokens(w) // Clear any existing (possibly invalid) tokens
				http.Redirect(w, r, "/login/", http.StatusFound)
				return
			}

			// Check for expired token
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorExpired != 0 {
					accessTokenClaims, err := myjwt.GetAccessClaims(accessToken)
					if err != nil {
						fmt.Println(err)
						myjwt.NullifyTokens(w)
						http.Redirect(w, r, "/login/", http.StatusFound)
						return
					}
					err = myjwt.IssueNewAccessToken(w, r, accessTokenClaims.UserID, accessTokenClaims.Username, accessTokenClaims.Role)
					if err != nil {
						fmt.Println("Error issuing new access token:", err)
						myjwt.NullifyTokens(w)
						http.Redirect(w, r, "/login/", http.StatusFound)
						return
					}
					fmt.Println("new Token")
					// procceed with the request
				} else if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
					fmt.Println("Signature validation failed")
					myjwt.NullifyTokens(w)
					http.Redirect(w, r, "/login/", http.StatusFound)
					return
				} else {
					fmt.Println("Other validation error")
					myjwt.NullifyTokens(w)
					http.Redirect(w, r, "/login/", http.StatusFound)
					return
				}
			}
		} else if !accessToken.Valid {
			fmt.Println("Token is invalid")
			myjwt.NullifyTokens(w)
			http.Redirect(w, r, "/login/", http.StatusFound)
			return
		} else {
			fmt.Println("Token is valid")
			// do nothing else
		}

		// Handle other paths or return a not found error
		next.ServeHTTP(w, r)
	})
}

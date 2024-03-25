package database

import (
	"database/sql"
	"fmt"
	"time"

	"main/models"

	_ "github.com/lib/pq"
)

var db *sql.DB

const (
	host     = "ep-polished-lab-a45rj9zc-pooler.us-east-1.aws.neon.tech"
	port     = 5432
	user     = "default"
	password = "DE3fRxcL6ugU"
	dbname   = "verceldb"
)

func InitDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println("Database connection error: ", err)
		return db, err
	}
	err = db.Ping()
	CleanExpiredOrRevokedTokens()
	return db, err
}

func CloseDB() (err error) {
	return db.Close()
}

func GetUserWithUsername(username string) (returnUser models.Users, err error) {
	sqlCommand := "SELECT * FROM users WHERE username = $1"
	row := db.QueryRow(sqlCommand, username)
	//  users: user_id, username, email ,hashed_password ,phone ,role, created_at
	err = row.Scan(&returnUser.UserID, &returnUser.Username, &returnUser.Email, &returnUser.HashedPassword, &returnUser.Phone, &returnUser.Role, &returnUser.CreatedAt)
	return returnUser, err
}

func AddUser(userArg models.Users) (err error) {
	sqlCommand := "INSERT INTO users(username, email, hashed_password, phone, role) VALUES($1, $2, $3, $4, $5)"
	_, err = db.Exec(sqlCommand, userArg.Username, userArg.Email, userArg.HashedPassword, userArg.Phone, userArg.Role)
	return err
}

func StoreRefreshToken(token models.RefreshToken) (err error) {
	sqlCommand := "INSERT INTO refresh_tokens(token_id, family_id, issued_at, expires_at, revokes_at) VALUES($1, $2, $3, $4, $5)"
	_, err = db.Exec(sqlCommand, token.TokenID, token.FamilyID, token.IssuedAt, token.ExpiresAt, token.RevokesAt)
	return err
}

func GetRefreshToken(tokenID string) (returnToken models.RefreshToken, err error) {
	sqlCommand := "SELECT * FROM refresh_tokens WHERE token_id = $1"
	row := db.QueryRow(sqlCommand, tokenID)
	err = row.Scan(&returnToken.TokenID, &returnToken.FamilyID, &returnToken.IssuedAt, &returnToken.ExpiresAt, &returnToken.RevokesAt)
	return
}

func RevokeTokenFamilyWithFamilyID(familyID string) (err error) {
	sqlCommand := "UPDATE refresh_tokens SET revokes_at = NOW() WHERE family_id = $1"
	_, err = db.Exec(sqlCommand, familyID)
	return err
}

func RevokeTokenWithTokenID(tokenID string) (err error) {
	sqlCommand := "UPDATE refresh_tokens SET revokes_at = NOW() + Interval '15 minutes' WHERE token_id = $1"
	_, err = db.Exec(sqlCommand, tokenID)
	return err
}

func RevokeTokenFamilyWithTokenID(tokenID string) error {
	// First, get the family_id associated with the given token_id
	familyID, err := GetFamilyIDFromTokenID(tokenID)
	if err != nil {
		// Handle error (return or log, depending on your error handling strategy)
		return err
	}

	// Now that we have the family_id, proceed to revoke all tokens in the family
	sqlCommand := "UPDATE refresh_tokens SET revokes_at = NOW() WHERE family_id = $1"
	_, err = db.Exec(sqlCommand, familyID)
	return err
}

func GetFamilyIDFromTokenID(tokenID string) (familyID string, err error) {
	sqlCommand := "SELECT family_id FROM refresh_tokens WHERE token_id = $1"
	err = db.QueryRow(sqlCommand, tokenID).Scan(&familyID)
	if err != nil {
		// Handle error (e.g., sql.ErrNoRows if token_id is not found)
		return "", err
	}
	return familyID, nil
}

func VerifyRefreshToken(tokenID string) (bool, error) {
	sqlCommand := "SELECT revokes_at FROM refresh_tokens WHERE token_id = $1"
	var revocationTime time.Time
	err := db.QueryRow(sqlCommand, tokenID).Scan(&revocationTime)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("no rows")
			return true, nil
		}
		return false, err
	}
	return revocationTime.Before(time.Now()), nil
}

func CleanExpiredOrRevokedTokens() error {
	sqlCommand := "DELETE FROM refresh_tokens WHERE expires_at < NOW()"
	_, err := db.Exec(sqlCommand)
	return err
}

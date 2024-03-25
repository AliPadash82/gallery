package main

import (
	"encoding/json"
	"fmt"
	"io"
	myjwt "main/MyJwt"
	"main/database"
	"main/models"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func endpointPassUsername(w http.ResponseWriter, r *http.Request) {
	accessTokenClaims, err := myjwt.ValidateAccessTokenAndGetClaims(r)
	if err != nil {
		fmt.Println("redirecting")
		http.Redirect(w, r, "/api/username", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	type Response struct {
		Username string `json:"username"`
		UserID   int    `json:"userid"`
	}
	data := Response{accessTokenClaims.Username, accessTokenClaims.UserID}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		http.Redirect(w, r, "/login/", http.StatusFound)
	}
	w.Write(jsonData)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	myjwt.NullifyTokens(w)
	http.Redirect(w, r, "/login/", http.StatusFound)
}

func homepageHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "homepage.html", nil)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	csrfToken, err := myjwt.GenerateRandomString()
	if err != nil {
		fmt.Println(err)
		return
	}
	myjwt.SetCSRFCookie(w, csrfToken)
	renderTemplate(w, "upload.html", struct{ CSRFToken string }{CSRFToken: csrfToken})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login.html", nil)
}

func galleryHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "gallery.html", nil)
}

func photosRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// Assuming db is your *sql.DB connection
	var photos []photo

	accessTokenCookie, err := myjwt.ValidateAccessTokenAndGetClaims(r)
	errHandle(w, err)

	rows, err := db.Query("SELECT title, filePath FROM photo where userid = $1", accessTokenCookie.UserID)
	errHandle(w, err)
	defer rows.Close()

	for rows.Next() {
		var p photo
		errHandle(w, rows.Scan(&p.Title, &p.FilePath))
		photos = append(photos, p)
	}

	errHandle(w, rows.Err())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photos)
}

func loginSubmitHandler(w http.ResponseWriter, r *http.Request) {
	errHandle(w, r.ParseForm())

	username := r.Form.Get("username")
	// password := r.Form.Get("password")

	newUser, err := database.GetUserWithUsername(username)
	errHandle(w, err)

	// errHandle(w, bcrypt.CompareHashAndPassword([]byte(newUser.HashedPassword), []byte(password)))
	fmt.Println(username, newUser.Username, newUser.UserID)

	myjwt.IssueTokens(w, newUser)

	http.Redirect(w, r, "/homepage/", http.StatusFound)
}

func signupSumbitHandler(w http.ResponseWriter, r *http.Request) {
	errHandle(w, r.ParseForm())
	username := r.FormValue("username")
	password := r.Form.Get("password")
	email := r.Form.Get("email")
	fmt.Println("new user: ", username, password, email)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	errHandle(w, err)
	newUser := models.Users{Username: username, Email: email, HashedPassword: string(hashedPassword), Role: "user"}
	err = database.AddUser(newUser)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/login/", http.StatusFound)
	} else {
		myjwt.IssueTokens(w, newUser)
		http.Redirect(w, r, "/homepage/", http.StatusFound)
	}
}

func submitPhotoHandler(w http.ResponseWriter, r *http.Request) {
	if !myjwt.ValidateCSRFToken(r) {
		myjwt.NullifyTokens(w)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"redirect": "/login/"})
		return
	}

	claims, err := myjwt.ValidateAccessTokenAndGetClaims(r)
	errHandle(w, err)
	// Parse the multipart form
	const _24K = (1 << 20) * 24
	if err := r.ParseMultipartForm(_24K); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	title := r.FormValue("title")
	description := r.FormValue("description")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a unique filename to avoid overwriting
	ext := filepath.Ext(header.Filename)
	uniqueName := uuid.New().String() + "-" + time.Now().Format("20060102150405") + title + ext
	filePath := filepath.Join("uploaded_images", uniqueName)

	// Save the file
	dst, err := os.Create(filePath)
	errHandle(w, err)
	defer dst.Close()

	// Copy the uploaded file to the destination file
	_, err = io.Copy(dst, file)
	errHandle(w, err)

	// Use the filePath in the SQL command
	sqlCommand := "INSERT INTO photo(userID, title, description, filePath) VALUES($1, $2, $3, $4)"

	_, err = db.Exec(sqlCommand, claims.UserID, title, description, filePath)
	errHandle(w, err)

	// Respond to the client
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Photo uploaded successfully"))
}

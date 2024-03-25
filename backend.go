package main

import (
	"database/sql"
	"fmt"
	myjwt "main/MyJwt"
	"main/database"
	"main/middleware"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type photo struct {
	PhotoID     int
	UserID      int
	Title       string `json:"title"`
	Description string `json:"description"`
	FilePath    string `json:"filePath"`
}

type users struct {
	userID   int
	username string
	password string
	email    string
}

var db *sql.DB

func main() {
	errHandle(nil, myjwt.InitJwt())
	var err error
	db, err = database.InitDB()
	errHandle(nil, err)
	defer func() {
		err = database.CloseDB()
		errHandle(nil, err)
	}()

	r := mux.NewRouter()

	// Serve static files directly without middleware
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", noCache(http.FileServer(http.Dir("assets")))))
	r.PathPrefix("/uploaded_images/").Handler(http.StripPrefix("/uploaded_images/", noCache(http.FileServer(http.Dir("uploaded_images")))))

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthHandler)
	api.HandleFunc("/username", endpointPassUsername)

	// Protected routes with Auth middleware
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthHandler)
	protected.HandleFunc("/gallery/", galleryHandler)
	protected.HandleFunc("/gallery/photos_retrieve", photosRetrieveHandler)
	protected.HandleFunc("/upload/", uploadHandler)
	protected.HandleFunc("/upload/upload", submitPhotoHandler)
	protected.HandleFunc("/homepage/", homepageHandler)
	protected.HandleFunc("/logout/", logoutHandler)

	// Login and signup routes without middleware
	r.HandleFunc("/login/", loginHandler)
	r.HandleFunc("/login/login_submit/", loginSubmitHandler)
	r.HandleFunc("/login/signup_submit/", signupSumbitHandler)

	fmt.Println("database is successfully connected.")
	errHandle(nil, http.ListenAndServe("127.0.0.1:8080", r))
}

// Assume other function definitions remain the same

package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"html/template"

	_ "github.com/lib/pq"
)

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    tmplPath := filepath.Join("templates", tmpl)
    parsedTmpl, err := template.ParseFiles(tmplPath)
    if err != nil {
        fmt.Println(err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    if err := parsedTmpl.Execute(w, data); err != nil {
        fmt.Println(err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}


func printUsers() {
	fmt.Println(db)
	rows, err := db.Query("SELECT userID, username, passwordHash, email FROM userr")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Iterate over the rows and print them
	fmt.Println("Rows in the userr table:")
	for rows.Next() {
		var userID int
		var username, email string
		var passwordHash []byte
		err = rows.Scan(&userID, &username, &passwordHash, &email)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("UserID: %d, Username: %s, Password: %x, Email: %s\n", userID, username, passwordHash, email)
	}

	// Check for errors after the loop
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func getColumnNamesForTable(tableName string) ([]string, error) {
	var columns []string
	query := `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = $1
		ORDER BY ordinal_position
	`
	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		columns = append(columns, columnName)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

// Prints all rows of the specified table.
func printTableRows(tableName string) {
	columnNames, err := getColumnNamesForTable(tableName)
	if err != nil {
		log.Fatalf("Failed to get column names: %v", err)
	}
	if len(columnNames) == 0 {
		log.Fatalf("No columns found for table: %s", tableName)
	}

	columnsCSV := strings.Join(columnNames, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s", columnsCSV, tableName)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query table rows: %v", err)
	}
	defer rows.Close()

	fmt.Printf("Rows in the %s table:\n", tableName)
	for rows.Next() {
		columnValues := make([]interface{}, len(columnNames))
		columnPointers := make([]interface{}, len(columnNames))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		for i, val := range columnValues {
			fmt.Printf("%s: %v, ", columnNames[i], val)
		}
		fmt.Println()
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error reading table rows: %v", err)
	}
}

func errHandle(w http.ResponseWriter, err error) error {
	if err != nil {
		if w == nil {
			panic(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return err
}

func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set headers to disable caching
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		w.Header().Set("Expires", "0")                                         // Proxies.
		h.ServeHTTP(w, r)
	})
}
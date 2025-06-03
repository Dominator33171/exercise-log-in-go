package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "./workouts.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Initialize router
	r := mux.NewRouter()

	// Home route with template
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/list.html")
		if err != nil {
			log.Printf("Template error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Create workout route
	r.HandleFunc("/workout/new", createWorkout(db)).Methods("POST")

	// Start server
	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func createWorkout(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			exercise := r.FormValue("exercise")
			durationStr := r.FormValue("duration")
			location := r.FormValue("location")
			description := r.FormValue("description")
			log.Printf("Form values: exercise=%s, duration=%s, location=%s, description=%s", exercise, durationStr, location, description)

			// Enhanced validation for #123
			if exercise == "" || durationStr == "" {
				log.Printf("Missing required fields")
				http.Error(w, "Exercise and duration are required", http.StatusBadRequest)
				return
			}

			duration, err := strconv.Atoi(durationStr)
			if err != nil {
				log.Printf("Invalid duration: %v", err)
				http.Error(w, "Invalid duration", http.StatusBadRequest)
				return
			}

			result, err := db.Exec("INSERT INTO workouts (exercise, duration, location, description) VALUES (?, ?, ?, ?)", exercise, duration, location, description)
			if err != nil {
				log.Printf("Database error: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			id, _ := result.LastInsertId()
			log.Printf("Inserted workout with ID: %d", id)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

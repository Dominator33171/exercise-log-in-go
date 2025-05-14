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

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

type Workout struct {
	ID          int
	Exercise    string
	Duration    int
	Location    string
	Description string
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./workouts.db")
	if err != nil {
		log.Fatal(err)
	}
	createTable := `
    CREATE TABLE IF NOT EXISTS workouts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        exercise TEXT,
        duration INTEGER,
        location TEXT,
        description TEXT
    );`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func listWorkouts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, exercise, duration, location, description FROM workouts")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		workouts := []Workout{}
		for rows.Next() {
			var wkt Workout
			if err := rows.Scan(&wkt.ID, &wkt.Exercise, &wkt.Duration, &wkt.Location, &wkt.Description); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			workouts = append(workouts, wkt)
		}
		tmpl.ExecuteTemplate(w, "list.html", workouts)
	}
}

func newWorkout(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "new.html", nil)
}

func createWorkout(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			exercise := r.FormValue("exercise")
			durationStr := r.FormValue("duration")
			location := r.FormValue("location")
			description := r.FormValue("description")
			log.Printf("Form values: exercise=%s, duration=%s, location=%s, description=%s", exercise, durationStr, location, description)
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

func showWorkout(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid workout ID", http.StatusBadRequest)
			return
		}
		var workout Workout
		err = db.QueryRow("SELECT id, exercise, duration, location, description FROM workouts WHERE id = ?", id).Scan(&workout.ID, &workout.Exercise, &workout.Duration, &workout.Location, &workout.Description)
		if err != nil {
			http.Error(w, "Workout not found", http.StatusNotFound)
			return
		}
		tmpl.ExecuteTemplate(w, "show.html", workout)
	}
}

func editWorkout(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid workout ID", http.StatusBadRequest)
			return
		}
		var workout Workout
		err = db.QueryRow("SELECT id, exercise, duration, location, description FROM workouts WHERE id = ?", id).Scan(&workout.ID, &workout.Exercise, &workout.Duration, &workout.Location, &workout.Description)
		if err != nil {
			http.Error(w, "Workout not found", http.StatusNotFound)
			return
		}
		tmpl.ExecuteTemplate(w, "edit.html", workout)
	}
}

func updateWorkout(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			vars := mux.Vars(r)
			id, err := strconv.Atoi(vars["id"])
			if err != nil {
				http.Error(w, "Invalid workout ID", http.StatusBadRequest)
				return
			}
			exercise := r.FormValue("exercise")
			durationStr := r.FormValue("duration")
			location := r.FormValue("location")
			description := r.FormValue("description")
			duration, err := strconv.Atoi(durationStr)
			if err != nil {
				http.Error(w, "Invalid duration", http.StatusBadRequest)
				return
			}
			_, err = db.Exec("UPDATE workouts SET exercise = ?, duration = ?, location = ?, description = ? WHERE id = ?", exercise, duration, location, description, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/workout/"+strconv.Itoa(id), http.StatusSeeOther)
		}
	}
}

func main() {
	db := initDB()
	defer db.Close()
	r := mux.NewRouter()
	r.HandleFunc("/", listWorkouts(db)).Methods("GET")
	r.HandleFunc("/workout/new", newWorkout).Methods("GET")
	r.HandleFunc("/workout/create", createWorkout(db)).Methods("POST")
	r.HandleFunc("/workout/{id:[0-9]+}", showWorkout(db)).Methods("GET")
	r.HandleFunc("/workout/{id:[0-9]+}/edit", editWorkout(db)).Methods("GET")
	r.HandleFunc("/workout/{id:[0-9]+}/update", updateWorkout(db)).Methods("POST")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

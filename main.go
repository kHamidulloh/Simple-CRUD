package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type CLS struct {
	Celsius int `json:"celsius"`
}

type DB struct {
	movies []Movie

	sync.RWMutex
}

func NewDB() *DB {
	return &DB{movies: make([]Movie, 0)}
}

type Movie struct {
	ID       string    `json:"id"`
	Isbn     string    `json:"isbn"`
	Title    string    `json:"title"`
	Director *Director `json:"director"`
}

type Director struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

func getMovies(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db.RLock()
		defer db.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(db.movies)
	}
}

func getMovie(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db.RLock()
		defer db.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		for _, item := range db.movies {
			if item.ID == params["id"] {
				json.NewEncoder(w).Encode(item)
				return
			}

		}
	}
}

func createMovie(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db.Lock()
		defer db.Unlock()

		w.Header().Set("Content-Type", "application/json")
		var movie Movie
		_ = json.NewDecoder(r.Body).Decode(&movie)
		movie.ID = strconv.Itoa(rand.IntN(100000))
		db.movies = append(db.movies, movie)
		json.NewEncoder(w).Encode(movie)
	}
}

func updateMovie(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db.Lock()
		defer db.Unlock()

		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)

		for index, item := range db.movies {
			if item.ID == params["id"] {
				db.movies = append(db.movies[:index], db.movies[index+1:]...)
				var movie Movie

				_ = json.NewDecoder(r.Body).Decode(&movie)
				movie.ID = strconv.Itoa(rand.IntN(100000))
				db.movies = append(db.movies, movie)
				json.NewEncoder(w).Encode(movie)
				return
			}
		}
	}
}

func deleteMovie(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db.Lock()
		defer db.Unlock()

		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		for index, item := range db.movies {
			if item.ID == params["id"] {
				db.movies = append(db.movies[:index], db.movies[index+1:]...)
				break
			}
		}
		json.NewEncoder(w).Encode(db.movies)
	}
}

func celsiusToFarangeyt(w http.ResponseWriter, r *http.Request) {
	var cls CLS
	err := json.NewDecoder(r.Body).Decode(&cls)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	far := (cls.Celsius * 9 / 5) + 32
	json.NewEncoder(w).Encode(far)
}

func main() {
	db := NewDB()
	r := mux.NewRouter()

	r.HandleFunc("/movies", getMovies(db)).Methods("GET")
	r.HandleFunc("/movies/{id}", getMovie(db)).Methods("GET")
	r.HandleFunc("/movies", createMovie(db)).Methods("POST")
	r.HandleFunc("/movies/{id}", updateMovie(db)).Methods("PUT")
	r.HandleFunc("/movies/{id}", deleteMovie(db)).Methods("DELETE")

	r.HandleFunc("/clstofar", celsiusToFarangeyt).Methods("POST")

	fmt.Printf("Starting server at the port 8000 ")
	log.Fatal(http.ListenAndServe(":8000", r))
}

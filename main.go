package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type id uuid.UUID

type user struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Biography string `json:"biography"`
}

type application struct {
	data map[id]user
}

func main() {
	app := application{data: make(map[id]user)}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/users", FindAll(app))
	mux.HandleFunc("GET /api/users/{id}", FindUserById(app))
	mux.HandleFunc("POST /api/users", CreateUser(app))
	mux.HandleFunc("PUT /api/users/{id}", UpdateUser(app))
	mux.HandleFunc("DELETE /api/users/{id}", DeleteUser(app))

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}

func FindAll(app application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users := make([]user, 0, len(app.data))

		for _, u := range app.data {
			users = append(users, u)
		}

		if err := json.NewEncoder(w).Encode(app.data); err != nil {
			http.Error(w, "could not encode response", http.StatusInternalServerError)
			return
		}
	}
}

func FindUserById(app application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parsed, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid ID", http.StatusBadRequest)
			return
		}

		user, ok := app.data[id(parsed)]
		if !ok {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, "could not encode response", http.StatusInternalServerError)
			return
		}
	}
}

func CreateUser(app application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u user

		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Invalid json", http.StatusBadRequest)
			return
		}

		id := id(uuid.New())

		app.data[id] = u

		w.WriteHeader(http.StatusCreated)
	}
}

func UpdateUser(app application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u user

		parsed, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid ID", http.StatusBadRequest)
			return
		}

		key := id(parsed)

		if _, ok := app.data[key]; !ok {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Invalid json", http.StatusBadRequest)
			return
		}

		app.data[key] = u

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)
	}
}

func DeleteUser(app application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parsed, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid ID", http.StatusBadRequest)
			return
		}

		key := id(parsed)

		if _, ok := app.data[key]; !ok {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		delete(app.data, key)
		w.WriteHeader(http.StatusNoContent)
	}
}

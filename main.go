package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type user struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Biography string `json:"biography"`
}

func (u user) validate() error {
	if strings.TrimSpace(u.FirstName) == "" {
		return errors.New("first_name is required")
	}
	if strings.TrimSpace(u.LastName) == "" {
		return errors.New("last_name is required")
	}
	if len(u.Biography) > 1000 {
		return errors.New("biography must be 1000 characters or fewer")
	}
	return nil
}

type store struct {
	mu   sync.RWMutex
	data map[string]user
}

func newStore() *store {
	return &store{data: make(map[string]user)}
}

func (s *store) list() []user {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]user, 0, len(s.data))
	for _, u := range s.data {
		users = append(users, u)
	}
	return users
}

func (s *store) get(key string) (user, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.data[key]
	return u, ok
}

func (s *store) create(u user) user {
	u.ID = uuid.NewString()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[u.ID] = u
	return u
}

func (s *store) update(key string, u user) (user, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		return user{}, false
	}
	u.ID = key
	s.data[key] = u
	return u, true
}

func (s *store) delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		return false
	}
	delete(s.data, key)
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func parseID(r *http.Request) (string, error) {
	parsed, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

func decodeUser(r *http.Request) (user, error) {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u user
	if err := dec.Decode(&u); err != nil {
		return user{}, err
	}
	return u, nil
}

func main() {
	s := newStore()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/users", FindAll(s))
	mux.HandleFunc("GET /api/users/{id}", FindUserById(s))
	mux.HandleFunc("POST /api/users", CreateUser(s))
	mux.HandleFunc("PUT /api/users/{id}", UpdateUser(s))
	mux.HandleFunc("DELETE /api/users/{id}", DeleteUser(s))

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("listening on :8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}

func FindAll(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.list())
	}
}

func FindUserById(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := parseID(r)
		if err != nil {
			http.Error(w, "invalid ID", http.StatusBadRequest)
			return
		}

		u, ok := s.get(key)
		if !ok {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		writeJSON(w, http.StatusOK, u)
	}
}

func CreateUser(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := decodeUser(r)
		if err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if err := u.validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		created := s.create(u)

		w.Header().Set("Location", "/api/users/"+created.ID)
		writeJSON(w, http.StatusCreated, created)
	}
}

func UpdateUser(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := parseID(r)
		if err != nil {
			http.Error(w, "invalid ID", http.StatusBadRequest)
			return
		}

		u, err := decodeUser(r)
		if err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if err := u.validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updated, ok := s.update(key, u)
		if !ok {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}

func DeleteUser(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := parseID(r)
		if err != nil {
			http.Error(w, "invalid ID", http.StatusBadRequest)
			return
		}

		if !s.delete(key) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

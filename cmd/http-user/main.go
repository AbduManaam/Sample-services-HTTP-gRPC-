package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// instanceID identifies this specific container instance.
// Set via INSTANCE_ID environment variable in docker-compose.
// Falls back to hostname (Docker sets this to container short ID)
// so it's always unique even without explicit config.
var instanceID string

func init() {
	instanceID = os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		var err error
		instanceID, err = os.Hostname()
		if err != nil {
			instanceID = "unknown"
		}
	}
}

// User represents a user entity
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// UserStore holds users in memory with thread-safe access
type UserStore struct {
	mu    sync.RWMutex
	users map[int]User
	idGen int
}

var userStore = &UserStore{
	users: make(map[int]User),
	idGen: 1,
}

func init() {
	userStore.users[1] = User{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Age: 28}
	userStore.users[2] = User{ID: 2, Name: "Bob Smith", Email: "bob@example.com", Age: 34}
	userStore.users[3] = User{ID: 3, Name: "Charlie Brown", Email: "charlie@example.com", Age: 42}
	userStore.idGen = 4
}

// healthHandler — instance_id added so LB verification is possible.
// When you hit this endpoint 6 times and see instance_id rotating
// between user-1, user-2, user-3, LB is confirmed working.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "ok",
		"service":     "user-service",
		"instance_id": instanceID,
	})
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	userStore.mu.RLock()
	users := make([]User, 0, len(userStore.users))
	for _, user := range userStore.users {
		users = append(users, user)
	}
	userStore.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users":       users,
		"count":       len(users),
		"instance_id": instanceID, // visible in every response for LB verification
		"served_by":   "user-service@" + instanceID,
	})
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/users/"), "/")
	idStr := pathParts[0]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user id"})
		return
	}

	userStore.mu.RLock()
	user, found := userStore.users[id]
	userStore.mu.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Wrap in envelope with instance_id so LB is visible
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":        user,
		"instance_id": instanceID,
	})
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var newUser struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid json body"})
		return
	}

	if newUser.Name == "" || newUser.Email == "" || newUser.Age <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "name, email, and age are required"})
		return
	}

	userStore.mu.Lock()
	id := userStore.idGen
	userStore.idGen++
	user := User{ID: id, Name: newUser.Name, Email: newUser.Email, Age: newUser.Age}
	userStore.users[id] = user
	userStore.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":        user,
		"instance_id": instanceID,
	})
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	headers := make(map[string][]string)
	for key, values := range r.Header {
		headers[key] = values
	}

	queryParams := make(map[string][]string)
	for key, values := range r.URL.Query() {
		queryParams[key] = values
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"instance_id":  instanceID, // most important field for LB debugging
		"method":       r.Method,
		"path":         r.URL.Path,
		"headers":      headers,
		"query_params": queryParams,
		"remote_addr":  r.RemoteAddr,
		"timestamp":    time.Now().Format(time.RFC3339),
		"forwarded_by": r.Header.Get("X-Gateway"),
	})
}

func router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/health":
		healthHandler(w, r)
	case path == "/users" && r.Method == http.MethodGet:
		listUsersHandler(w, r)
	case path == "/users" && r.Method == http.MethodPost:
		createUserHandler(w, r)
	case strings.HasPrefix(path, "/users/") && r.Method == http.MethodGet:
		getUserHandler(w, r)
	case path == "/debug":
		debugHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}

	http.HandleFunc("/", router)

	log.Printf("User Service starting | instance_id=%s | port=%s", instanceID, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Order represents an order entity
type Order struct {
	ID       int     `json:"id"`
	UserID   int     `json:"user_id"`
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Status   string  `json:"status"`
}

// OrderStore holds orders in memory with thread-safe access
type OrderStore struct {
	mu     sync.RWMutex
	orders map[int]Order
	idGen  int
}

var orderStore = &OrderStore{
	orders: make(map[int]Order),
	idGen:  1,
}

// Initialize with mock data
func init() {
	orderStore.orders[1] = Order{ID: 1, UserID: 1, Product: "Laptop", Quantity: 1, Price: 999.99, Status: "shipped"}
	orderStore.orders[2] = Order{ID: 2, UserID: 2, Product: "Mouse", Quantity: 2, Price: 29.99, Status: "delivered"}
	orderStore.orders[3] = Order{ID: 3, UserID: 1, Product: "Keyboard", Quantity: 1, Price: 79.99, Status: "pending"}
	orderStore.idGen = 4
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "order-service",
	})
}

// Get all orders
func listOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orderStore.mu.RLock()
	orders := make([]Order, 0, len(orderStore.orders))
	for _, order := range orderStore.orders {
		orders = append(orders, order)
	}
	orderStore.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"orders": orders,
		"count":  len(orders),
	})
}

// Get single order by ID
func getOrderHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/orders/"), "/")
	idStr := pathParts[0]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid order id"})
		return
	}

	orderStore.mu.RLock()
	order, found := orderStore.orders[id]
	orderStore.mu.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "order not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// Create new order
func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var newOrder struct {
		UserID   int     `json:"user_id"`
		Product  string  `json:"product"`
		Quantity int     `json:"quantity"`
		Price    float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid json body"})
		return
	}

	if newOrder.UserID <= 0 || newOrder.Product == "" || newOrder.Quantity <= 0 || newOrder.Price <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "user_id, product, quantity, and price are required and must be positive"})
		return
	}

	orderStore.mu.Lock()
	id := orderStore.idGen
	orderStore.idGen++
	order := Order{ID: id, UserID: newOrder.UserID, Product: newOrder.Product, Quantity: newOrder.Quantity, Price: newOrder.Price, Status: "pending"}
	orderStore.orders[id] = order
	orderStore.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// Debug endpoint - echoes request details
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
		"method":       r.Method,
		"path":         r.URL.Path,
		"headers":      headers,
		"query_params": queryParams,
		"remote_addr":  r.RemoteAddr,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
}

// Router handler
func router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/health":
		healthHandler(w, r)
	case path == "/orders" && r.Method == http.MethodGet:
		listOrdersHandler(w, r)
	case path == "/orders" && r.Method == http.MethodPost:
		createOrderHandler(w, r)
	case strings.HasPrefix(path, "/orders/") && r.Method == http.MethodGet:
		getOrderHandler(w, r)
	case path == "/debug":
		debugHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}
}

func main() {
	http.HandleFunc("/", router)

	port := ":9002"
	log.Printf("Order Service starting on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

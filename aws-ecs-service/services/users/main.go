package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func main() {

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("service users version " + os.Getenv("API_VERSION")))
	})

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		type User struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		users := []User{
			{ID: 1, Name: "John Doe", Email: "jhon@gmail.com"},
			{ID: 2, Name: "Jane Doe", Email: "jane@gmail.com"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})

	r.Get("/service-products", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(os.Getenv("SERVICE_PRODUCTS") + "/products")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(body)
	})

	r.Get("/stress-test", func(w http.ResponseWriter, r *http.Request) {
		limit := 1000000
		for i := 2; i <= limit; i++ {
			if isPrime(i) {
				fmt.Printf("%d is a prime number\n", i)
			}
		}

		w.Write([]byte("Stress test completed"))
	})

	http.ListenAndServe(":3000", r)
}

package router

import (
	"database/sql"
	"net/http"
	"server/controllers"
	"server/internal"
	"server/models"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/cors"
)

func GetRouter(
	dbInstance *sql.DB,
	jwtService *internal.JWTService,
) *chi.Mux {
	const apiVersion = "/api/v1"

	userController := controllers.User{
		UserService: models.UserService{
			DB: dbInstance,
		},
		JWTService: *jwtService,
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(c.Handler)

	r.Use(httprate.LimitByIP(200, time.Minute))

	r.Get(apiVersion+"/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v1.0.0"))
	})

	r.Post(apiVersion+"/login", func(w http.ResponseWriter, r *http.Request) {
		userController.Login(w, r)
	})

	return r
}

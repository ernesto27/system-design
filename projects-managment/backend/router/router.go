package router

import (
	"database/sql"
	"net/http"
	"server/controllers"
	"server/internal"
	"server/models"
	"server/response"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/cors"
)

func AuthMiddleware(jwtService *internal.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				response.NewWithoutData().WithMessage("Missing authorization token").Unauthorized(w)
				return
			}
			if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
				tokenString = tokenString[7:]
			}

			userID, _, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				response.NewWithoutData().WithMessage("Invalid token").Unauthorized(w)
				return
			}

			ctx := r.Context()
			ctx = internal.SetUserContext(ctx, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

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

	projectController := controllers.Project{
		ProjectService: models.ProjectService{
			DB: dbInstance,
		},
	}

	roleController := controllers.RoleController{
		RoleService: models.RoleService{
			DB: dbInstance,
		},
	}

	projectStatusController := controllers.ProjectStatusController{
		ProjectStatusService: models.ProjectStatusService{
			DB: dbInstance,
		},
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

	r.Get(apiVersion+"/project-statuses", func(w http.ResponseWriter, r *http.Request) {
		projectStatusController.GetAllProjectStatuses(w, r)
	})

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(jwtService))

		r.Post(apiVersion+"/projects", func(w http.ResponseWriter, r *http.Request) {
			projectController.Create(w, r)
		})

		r.Get(apiVersion+"/projects", func(w http.ResponseWriter, r *http.Request) {
			projectController.GetAll(w, r)
		})

		r.Get(apiVersion+"/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
			projectController.GetByID(w, r)
		})

		r.Put(apiVersion+"/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
			projectController.Update(w, r)
		})

		r.Get(apiVersion+"/roles", func(w http.ResponseWriter, r *http.Request) {
			roleController.GetRoles(w, r)
		})
	})

	return r
}

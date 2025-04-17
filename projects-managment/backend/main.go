package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"server/db"
	"server/internal"
	"server/router"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

//go:embed migrations/*.sql templates/*.html
var embedFS embed.FS

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Database
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	port := os.Getenv("DB_PORT")
	database := os.Getenv("DB_DATABASE")

	dbInstance, err := db.NewPostgres(host, user, password, port, database, "disable")
	if err != nil {
		panic(err)
	}

	err = db.RunMigrations(dbInstance.Db, embedFS)
	if err != nil {
		panic(err)
	}

	jwtService := internal.NewJWTService(os.Getenv("JWT_SECRET_KEY"))
	r := router.GetRouter(dbInstance.Db, jwtService)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("listen: %s\n", err)
		}
	}()

	fmt.Println("Starting the server on port 8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %s\n", err)
	}

	fmt.Println("Server exiting")
}

package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("invalid file"))
		return
	}
	defer file.Close()

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("file too big"))
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(file); err != nil {
		fmt.Println(err)
		w.Write([]byte("invalid file"))
		return
	}

	hash, err := generateRandomToken(32)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	myS3 := NewS3(os.Getenv("AWS_S3_REGION"), os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET"), os.Getenv("AWS_S3_BUCKET"))
	err = myS3.Upload(buf, fileHeader, hash)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	err = db.CreateFile(File{
		Name: fileHeader.Filename,
		Size: fileHeader.Size,
		Hash: hash,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte("file ok"))
}

func getFileHandler(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	myS3 := NewS3(os.Getenv("AWS_S3_REGION"), os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET"), os.Getenv("AWS_S3_BUCKET"))
	body, err := myS3.Get(hash)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(body)
}

var db *Mysql

func main() {
	err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	db, err = NewMysql(os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Post("/upload", uploadHandler)
	r.Get("/file/{hash}", getFileHandler)

	http.ListenAndServe(":3000", r)
}

func loadEnvConfig() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}

func generateRandomToken(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	hasher := sha256.New()
	hasher.Write(bytes)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	err := db.CreateFile(File{
		Name: r.FormValue("name"),
		Size: r.FormValue("size"),
		Hash: r.FormValue("hash"),
	})
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	return

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("invalid file"))
		return
	}
	defer file.Close()

	err = r.ParseMultipartForm(10 << 20) // 10MB maximum file size
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

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_S3_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET"), ""),
	}))

	svc := s3.New(sess)

	timeout := 1000 * time.Second
	ctx := context.Background()
	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}

	if cancelFn != nil {
		defer cancelFn()
	}

	_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:             aws.String(os.Getenv("AWS_S3_BUCKET")),
		Body:               bytes.NewReader(buf.Bytes()),
		Key:                aws.String(fileHeader.Filename),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
			w.Write([]byte("upload canceled due to timeout"))
			return
		} else {
			fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
			w.Write([]byte("failed to upload object"))
			return
		}
	}

	w.Write([]byte("file uploaded"))
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

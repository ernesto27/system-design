package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB maximum file size
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file from form data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes := bytes.Buffer{}
	_, err = io.Copy(&fileBytes, file)
	if err != nil {
		http.Error(w, "Failed to read file content", http.StatusInternalServerError)
		return
	}

	// Upload to bucket AWS
	err = uploadToS3(fileHeader.Filename, bytes.NewReader(fileBytes.Bytes()))
	if err != nil {
		http.Error(w, "Failed to upload file to S3", http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

func uploadToS3(name string, fileContent io.ReadSeeker) error {
	bucket := os.Getenv("AWS_S3_BUCKET")
	key := os.Getenv("AWS_S3_ACCESS_KEY")
	secret := os.Getenv("AWS_S3_SECRET")
	timeout := 1000 * time.Second

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
	}))

	svc := s3.New(sess)

	ctx := context.Background()
	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}

	if cancelFn != nil {
		defer cancelFn()
	}

	_, err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:             aws.String(bucket),
		Body:               fileContent,
		Key:                aws.String(name),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
			return err
		} else {
			fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
			return err
		}
	}

	return nil

}

func server() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})
	http.HandleFunc("/upload", uploadFileHandler)
	http.ListenAndServe(":8080", nil)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	server()
	return

	// cmd := exec.Command("ffmpeg", "-i", "video.mp4", "-vf", "scale=1280:720", "output.mp4")
	// err := cmd.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// return

	// var timeout time.Duration

	// bucket := os.Getenv("AWS_S3_BUCKET")
	// key := os.Getenv("AWS_S3_ACCESS_KEY")
	// secret := os.Getenv("AWS_S3_SECRET")
	// timeout = 1000 * time.Second

	// sess := session.Must(session.NewSession(&aws.Config{
	// 	Region:      aws.String("us-west-2"),
	// 	Credentials: credentials.NewStaticCredentials(key, secret, ""),
	// }))

	// svc := s3.New(sess)

	// ctx := context.Background()
	// var cancelFn func()
	// if timeout > 0 {
	// 	ctx, cancelFn = context.WithTimeout(ctx, timeout)
	// }

	// if cancelFn != nil {
	// 	defer cancelFn()
	// }

	// file, err := os.Open("image.jpg")
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, "Error opening file:", err)
	// 	return
	// }
	// defer file.Close()

	// o, err := svc.GetObject(&s3.GetObjectInput{
	// 	Bucket: aws.String(bucket),
	// 	Key:    aws.String("image1.jpg"),
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// file, err = os.Create("download1.jpg")
	// if err != nil {
	// 	fmt.Println("Error creating file:", err)
	// 	return
	// }
	// defer file.Close()

	// _, err = io.Copy(file, o.Body)
	// if err != nil {
	// 	fmt.Println("Error copying data to file:", err)
	// 	return
	// }

	// return

	// // Uploads the object to S3. The Context will interrupt the request if the
	// // timeout expires.
	// _, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
	// 	Bucket:             aws.String(bucket),
	// 	Body:               file,
	// 	Key:                aws.String("image1.jpg"),
	// 	ContentDisposition: aws.String("attachment"),
	// })
	// if err != nil {
	// 	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
	// 		// If the SDK can determine the request or retry delay was canceled
	// 		// by a context the CanceledErrorCode error code will be returned.
	// 		fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
	// 	} else {
	// 		fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
	// 	}
	// 	os.Exit(1)
	// }

	// fmt.Printf("successfully uploaded file to %s/%s\n", bucket, key)
}

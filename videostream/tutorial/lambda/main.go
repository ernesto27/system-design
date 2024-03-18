package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Hello lambda",
	}
	return response, nil
}

func convertFile(inputFile string, outputFile string) error {
	// ffmpeg -i world.mp4 -vcodec h264 -acodec mp2 output.mp4
	// out, err := exec.Command("ffmpeg", "-i", inputFile, "-vcodec", "h264", "-acodec", "mp2", outputFile).Output()
	out, err := exec.Command("ffmpeg", "-i", inputFile, "-vf", "scale=640:480", outputFile).Output()

	if err != nil {
		fmt.Printf("%s", err)
		return err
	}

	fmt.Println("Command Successfully Executed")
	res := string(out[:])
	fmt.Println(res)
	return nil
}

func handlerTriggerS3Bucket(ctx context.Context, s3Event events.S3Event) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bucketOutput := os.Getenv("AWS_S3_BUCKET")
	key := os.Getenv("AWS_ACCESS_KEY")
	secret := os.Getenv("AWS_SECRET_KEY")
	region := os.Getenv("AWS_REGION")
	timeout := 1000 * time.Second

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
	}))

	svc := s3.New(sess)
	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}

	if cancelFn != nil {
		defer cancelFn()
	}

	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.URLDecodedKey

		obj, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Printf("error getting head of object %s/%s: %s", bucket, key, err)
			return err
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, obj.Body); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading file:", err)
			return err
		}

		inputFile := "/tmp/input-" + key
		outputFile := "/tmp/output-" + key
		err = os.WriteFile(inputFile, buf.Bytes(), 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		err = convertFile(inputFile, outputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		fileToUpload, err := os.ReadFile(outputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket:             aws.String(bucketOutput),
			Body:               bytes.NewReader(fileToUpload),
			Key:                aws.String(key),
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
	}

	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	lambda.Start(handlerTriggerS3Bucket)
}

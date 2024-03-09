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
	// convertFile()
	response := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       *getBucketList(),
	}
	return response, nil
}

func getBucketList() *string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// bucket := os.Getenv("AWS_S3_BUCKET")
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

	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		fmt.Printf("Couldn't list buckets for your account. Here's why: %v\n", err)
		return nil
	}

	fmt.Println(result)
	return result.Buckets[0].Name
}

func convertFile() {
	outputVideo := "/tmp/output.mp4"
	out, err := exec.Command("ffmpeg", "-i", "/tmp/video.mp4", "-vf", "scale=640:480", outputVideo).Output()

	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	fmt.Println("Command Successfully Executed")
	res := string(out[:])
	fmt.Println(res)

	// out, err := exec.Command("ffmpeg", "-help").Output()
	// if err != nil {
	// 	fmt.Printf("%s", err)
	// 	return
	// }
	// fmt.Println(string(out[:]))
}

func handlerTriggerS3Bucket(ctx context.Context, s3Event events.S3Event) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// bucket := os.Getenv("AWS_S3_BUCKET")
	key := os.Getenv("AWS_S3_ACCESS_KEY")
	secret := os.Getenv("AWS_S3_SECRET")
	timeout := 1000 * time.Second

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
	}))

	svc := s3.New(sess)

	// ctx := context.Background()
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

		headOutput, err := svc.HeadObject(&s3.HeadObjectInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			log.Printf("error getting head of object %s/%s: %s", bucket, key, err)
			return err
		}
		log.Printf("successfully retrieved %s/%s of type %s", bucket, key, *headOutput.ContentType)

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, obj.Body); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading file:", err)
			return err
		}

		filename := "/tmp/video.mp4"
		// Write the bytes to the file
		err = os.WriteFile(filename, buf.Bytes(), 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		convertFile()

		fileToUpload, err := os.ReadFile("/tmp/output.mp4")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket:             aws.String("output-bucket-111"),
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
	//lambda.Start(handler)
	lambda.Start(handlerTriggerS3Bucket)
	//convertFile()
}

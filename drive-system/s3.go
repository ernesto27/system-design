package main

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type MyS3 struct {
	Region    string
	AccessKey string
	Secret    string
	Bucket    string
	client    *s3.S3
}

func NewS3(region, accessKey, secret, bucket string) *MyS3 {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secret, ""),
	}))

	return &MyS3{
		Region:    region,
		AccessKey: accessKey,
		Secret:    secret,
		Bucket:    bucket,
		client:    s3.New(sess),
	}
}

func (myS3 *MyS3) Get(fileName string) ([]byte, error) {
	result, err := myS3.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(myS3.Bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return nil, err
	}

	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

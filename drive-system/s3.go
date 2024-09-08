package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
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

func (myS3 *MyS3) Upload(buf *bytes.Buffer, fileHeader *multipart.FileHeader, hash string) error {
	timeout := 1000 * time.Second
	ctx := context.Background()
	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}

	if cancelFn != nil {
		defer cancelFn()
	}

	ext := filepath.Ext(fileHeader.Filename)

	_, err := myS3.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:             aws.String(myS3.Bucket),
		Body:               bytes.NewReader(buf.Bytes()),
		Key:                aws.String(hash + ext),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			return fmt.Errorf("upload canceled due to timeout, %v", err)
		} else {
			return err
		}
	}

	return nil
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

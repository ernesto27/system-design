TODO

- Explain project
- Diagram architecture 
- Requirements
- Obtener access, key secret key aws, permisos
- Crear bucket input
- Crear role, execution role
- Crear security group port 80
- Crer ecs project
- Create docker image,  upload to registry
- Create task definition
- Create service
- Test version app endpint
- Test endponint save file on bucket s3
- Crear bucket output
- Crear lambda with hello world
- Crear trigger aws bucket lambda
- Use ffmpeg to convert video and save on another bucket
- ECS use load balancer


## 1. Introduction
Hi, in this tutorial we are going to create a project that allow us to upload a video using a endpoint HTTP and then convert it to another resolution format, think of this as a process that a video service like youtube o vimeo does when an user upload a video on their platform,  in order to accomplish that we are going to use the following AWS services:

S3: Simple Storage Service
We will use this service to store the videos in the original and convert resolution format,  we are going to create two different buckets for each purpose.

Elastic Container Registry:
This service is for store our docker images for the tutorial.


ECS: Elastic Container Service
We will deploy our API http using this AWS service,  this allows us to run our application in a scalable way, using container without going into the complexity of managing the infrastructure by hand.

Lambda: Serverless functions.
We will use this service to trigger a event/function when a video in the original format is uploaded to the bucket,  this will get the video and convert it to another resolution after that save that on a different bucket.


### Diagram architecture

![Architecture](./diagram.png)


## Requirements

- AWS account - https://aws.amazon.com/
- Docker - https://docs.docker.com/engine/install/
- Go - https://go.dev/doc/install
- AWS CLI - https://aws.amazon.com/cli/

We need to create and account on AWS in order to follow this tutorial,  
Our services, API will use golang, so we need that on our machine.
We use docker to create our images,  that images will be upload to ECR,  and run on ECS and Lambda services.
Besides we will create most of the AWS resources using the web dashboard,  we must install the AWS cli to upload the docker image.

# Create user on AWS IAM 

We need to create a user on our AWS account, this user will have an access key and a secret key that we will need later to connect to the AWS services that we will create,  also we should limit the permissions and acces to this user to only the services that they need, this is very important for security reasons and for prevent any unwanted action on our account.

On the search box of AWS dashboard search for IAM and click on the first option that appears. 
On Access management sidebar click on user section,  after that click on Create User button on main dashboard section.

On name put a name "aws-tutorail" or anything you want is not important at this point, after click on Next.

We now need to add the permissions that this user will have, we are going to add manually this settings, so check that option on the dasboard.

![user permisions](./01-user-permissions.png)

We need to add the following permissions policies:

#### AmazonS3FullAccess
This is required for our API services to upload files on the bucket

#### AmazonEC2ContainerRegistryFullAccess
This is required to upload our docker images to the ECR service

Use the search box and select both policies, after that click on next.

![policies](./policies.png)

### Create acces key
After succesfuly create the user, enter on detail section of the user previously created.

Go to

**Security credentials -> Create access key**

On use cases,  select other ( AWS list other options and alternatives that we do not care at this moment ), click on Next,  you can put a description if you 
want,  after that click on Create access key.

After that you see the access key and private key on the dashboard,  copy that on some secure place of your own, this is because is the only and last time that AWS shows your private key.


# Create S3 buckets

Our application needs two buckets, one for the original file uploads and another for the convert video files,  so go to the aws dashboard and search for S3,  click on the first option that appears.


Click on create bucket, a new form configuration appears, most of the options should work fine on default values,  but we need to change the following:

AWS Region: You can selecte wherever you want, in this tutorial we are going to use us-east-2 Oregon region.

Bucket name: this must be a unique name, for this tutorial choose something that has the input name on it,  for example "tutorial-input", this semantic helps later when we have to create a trigger on a lambda function.

We mantain the select option for block all public access,  we are going to use the bucket only for our application,  so we do not need to make it public.

After that click on create bucket.

![user permisions](./02-bucket.png)

For the output bucket, repeat the same steps as before, only remember to change the name of the bucket to something like "tutorial-output".


# Create API service

For this tutorial we use golang to create our API,  we will create two endpoints, one for obtain a file from a client and save it on the input bucket, and anothe for show the current version of the application.

Start by creating a new project folder on your machine, name it ecs-tutorial,  inside that folder the following command.

```bash
go mod init ecs-tutorial
```

### Install dependencies

```bash
go get github.com/aws/aws-sdk-go
```
This is the AWS sdk that allow us to connect to different services,  in this case we are going to connect to S3 storage service

```bash
go get github.com/joho/godotenv
```
This library is util for use and .env file on our project,  we are going to use it for store our AWS credentials, version, etc.

```bash
go get github.com/go-chi/chi
```
This is a lightweight router HTTP, we are going to use it for create our API endpoints.

Create a .env file with this content.

```
AWS_S3_BUCKET=your-bucket-name
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=your-region
PORT=8080
```
Replaces the values with your own, the bucket the we use in this service is the "input" bucket that we created before.

Create a main.go  file with the following content.

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
)

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("upload"))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("version 0.0.1"))
	})
	r.Post("/upload", uploadFileHandler)

	http.ListenAndServe(":"+os.Getenv("PORT"), r)
}

```

On main we first obtain the environment variables defined on .env, if that fails panic and finish program, 
after create a chi router,  we use the middleware logger for log request info, data on the console,  
lastly create two endpoints, one for the root path that show the version of the application, and another for the upload file,
the funcion uploadFileHandler is return a hardcoded string "upload" for now, later we will add the logic to save the file on the S3 bucket.

Run server

```bash
go run main.go
```

Test using curl

```bash
curl http://localhost:8080/upload
```














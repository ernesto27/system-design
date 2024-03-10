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

## Create user on AWS IAM 

We need to create a user on our AWS account, 
this user will have a access key and secret key  we will need that later to connect to the AWS services that we will create,  also we must limit the permissions and acces to this user to only the services that they need, this is very important for security reasons and for prevent any unwanted action on our account.






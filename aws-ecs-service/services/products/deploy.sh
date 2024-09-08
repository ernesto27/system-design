#!/bin/bash

if [ $# -ne 3 ]; then
    echo "Usage: deploy.sh <registry_name> <image_name> <region>"
    exit 1
fi

REGISTRY_URI=$1
REGISTRY_NAME=$2
REGION=$3

aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $REGISTRY_URI

docker build -t $REGISTRY_NAME .

docker tag $IMAGE_NAME $REGISTRY_URI/$REGISTRY_NAME

docker push $REGISTRY_URI/$REGISTRY_NAME



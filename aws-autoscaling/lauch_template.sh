#!/bin/bash
set -e

# Update packages
apt-get update -y
apt-get upgrade -y

# Install Docker + AWS CLI dependencies
apt-get install -y docker.io jq curl unzip

systemctl start docker
systemctl enable docker
usermod -aG docker ubuntu

sudo apt install -y unzip

# Install AWS CLI v2 (Ubuntu 24 does not always have the latest CLI preinstalled)
cd /tmp
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o awscliv2.zip
unzip awscliv2.zip
sudo ./aws/install

# Variables
REGION="us-east-1"                   # adjust to your region
REPO_NAME="node-cpu"               # adjust to your ECR repo name
IMAGE_TAG="latest"                       # adjust to your tag

ACCOUNT_ID="383660184915"   

# Full ECR image URI
IMAGE_URI="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO_NAME}:latest"

# Authenticate Docker to ECR (using instance role + AWS CLI)
aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com

# Pull and run the container
docker pull $IMAGE_URI
docker run -d -p 8080:3000 --restart always --name myapp $IMAGE_URI


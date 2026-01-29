#!/bin/bash
set -e

AWS_REGION="us-east-1"
AWS_ACCOUNT="383660184915"
REPO_NAME="bunjs"
IMAGE_TAG="${1:-latest}"

ECR_URL="${AWS_ACCOUNT}.dkr.ecr.${AWS_REGION}.amazonaws.com"
FULL_IMAGE="${ECR_URL}/${REPO_NAME}:${IMAGE_TAG}"

echo "==> Logging in to ECR..."
aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${ECR_URL}

echo "==> Building image..."
docker build -t ${REPO_NAME} .

echo "==> Tagging image..."
docker tag ${REPO_NAME}:latest ${FULL_IMAGE}

echo "==> Pushing to ECR..."
docker push ${FULL_IMAGE}

echo "==> Done! Image pushed to:"
echo "    ${FULL_IMAGE}"

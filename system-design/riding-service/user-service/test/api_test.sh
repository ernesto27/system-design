#!/bin/bash

BASE_URL="http://localhost:8080"
TEST_EMAIL="test@example.com"
TEST_PASSWORD="password123"

# Function to check if jq is installed
check_jq() {
    if ! command -v jq &> /dev/null; then
        echo "jq is not installed. Please install it first."
        exit 1
    fi
}

check_jq

# Register new user
echo "Testing registration..."
REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${TEST_EMAIL}\",
    \"password\": \"${TEST_PASSWORD}\",
    \"name\": \"Test User\"
  }")


echo "Registration response: ${REGISTER_RESPONSE}"

# Wait a moment before login
sleep 1

# Login
echo "Testing login..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${TEST_EMAIL}\",
    \"password\": \"${TEST_PASSWORD}\"
  }")

echo "Login response: ${LOGIN_RESPONSE}"

# Extract token
TOKEN=$(echo "${LOGIN_RESPONSE}" | jq -r '.token')

if [ "${TOKEN}" == "null" ] || [ -z "${TOKEN}" ]; then
    echo "Failed to get token"
    exit 1
fi

echo "Token obtained: ${TOKEN}"
echo -e "\n"

# Get profile
echo "Testing profile retrieval..."
curl -s -X GET "${BASE_URL}/api/profile" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" | jq .
echo -e "\n"

#!/bin/bash

BASE_URL="http://localhost:8080"

echo "Testing Webhook API Endpoint"
echo "============================="

# Test 1: Valid webhook request
echo -e "\n1. Testing valid webhook request:"
curl -X POST $BASE_URL/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "id": "evt_123",
    "source": "shopify",
    "type": "order.created",
    "data": {
      "order_id": "12345",
      "customer": "john@example.com",
      "amount": 99.99
    }
  }' \
  -w "\nStatus Code: %{http_code}\n"

# Test 2: Invalid request - missing required fields
echo -e "\n2. Testing invalid request (missing source):"
curl -X POST $BASE_URL/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "id": "evt_124",
    "type": "order.created",
    "data": {
      "order_id": "12346"
    }
  }' \
  -w "\nStatus Code: %{http_code}\n"

echo -e "\nTest completed!"
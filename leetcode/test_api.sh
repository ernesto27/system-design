#!/bin/bash

echo "🚀 Testing LeetCode API Endpoints"
echo "=================================="
echo ""

BASE_URL="http://localhost:8080"

echo "1. Testing GET /problems (all problems)"
echo "---------------------------------------"
curl -s "$BASE_URL/problems" | jq '.'
echo ""

echo "2. Testing GET /problems with start parameter"
echo "--------------------------------------------"
curl -s "$BASE_URL/problems?start=1" | jq '.'
echo ""

echo "3. Testing GET /problems with start and end parameters"
echo "----------------------------------------------------"
curl -s "$BASE_URL/problems?start=0&end=2" | jq '.'
echo ""

echo "4. Testing GET /problems with only end parameter"
echo "-----------------------------------------------"
curl -s "$BASE_URL/problems?end=1" | jq '.'
echo ""

echo "5. Testing GET /problems/:problem_id (ID: 1)"
echo "-------------------------------------------"
curl -s "$BASE_URL/problems/1" | jq '.'
echo ""

echo "6. Testing GET /problems/:problem_id (ID: 3)"
echo "-------------------------------------------"
curl -s "$BASE_URL/problems/3" | jq '.'
echo ""

echo "7. Testing GET /problems/:problem_id with invalid ID"
echo "---------------------------------------------------"
curl -s "$BASE_URL/problems/999" | jq '.'
echo ""

echo "8. Testing GET /problems/:problem_id with non-numeric ID"
echo "-------------------------------------------------------"
curl -s "$BASE_URL/problems/abc" | jq '.'
echo ""

echo "9. Testing GET /problems with invalid start parameter"
echo "----------------------------------------------------"
curl -s "$BASE_URL/problems?start=abc" | jq '.'
echo ""

echo "10. Testing GET /problems with invalid end parameter"
echo "---------------------------------------------------"
curl -s "$BASE_URL/problems?end=xyz" | jq '.'
echo ""

echo "11. Testing POST /problems/:problem_id/submission (valid)"
echo "--------------------------------------------------------"
curl -s -X POST "$BASE_URL/problems/1/submission" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "1",
    "code": "def two_sum(nums, target):\n    for i in range(len(nums)):\n        for j in range(i+1, len(nums)):\n            if nums[i] + nums[j] == target:\n                return [i, j]",
    "language": "python"
  }' | jq '.'
echo ""

echo "12. Testing POST /problems/:problem_id/submission (invalid problem)"
echo "------------------------------------------------------------------"
curl -s -X POST "$BASE_URL/problems/999/submission" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "1",
    "code": "some code",
    "language": "python"
  }' | jq '.'
echo ""

echo "13. Testing POST /problems/:problem_id/submission (invalid user)"
echo "---------------------------------------------------------------"
curl -s -X POST "$BASE_URL/problems/1/submission" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "999",
    "code": "some code",
    "language": "python"
  }' | jq '.'
echo ""

echo "14. Testing POST /problems/:problem_id/submission (missing fields)"
echo "-----------------------------------------------------------------"
curl -s -X POST "$BASE_URL/problems/1/submission" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "1"
  }' | jq '.'
echo ""

echo "✅ API Testing Complete!"
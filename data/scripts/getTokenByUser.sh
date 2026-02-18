#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
EMAIL="test_1770841633712023786@example.com"
PASSWORD="password123"
NAME="Test User"

echo "Using email: $EMAIL"

echo -e "\nStep 1: Logging in..."
LOGIN_RES=$(curl -s -X POST "$API_URL/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\", \"password\":\"$PASSWORD\"}")
echo "Login Response: $LOGIN_RES"

TOKEN=$(echo $LOGIN_RES | grep -oP '"token":"\K[^"]+')

if [ -z "$TOKEN" ]; then
    echo "ERROR: Failed to get token from response."
    echo "Make sure the server is healthy and the credentials are correct."
    exit 1
fi

echo "Login successful. Token acquired."
echo "Token: $TOKEN"

#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
EMAIL="user_test@example.com"
PASSWORD="password123"
NAME="Test User"

echo "Using email: $EMAIL"

echo "Step 1: Registering user..."
REG_RES=$(curl -s -X POST "$API_URL/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\", \"password\":\"$PASSWORD\", \"name\":\"$NAME\"}")
echo "Registration Response: $REG_RES"

echo -e "\nStep 2: Logging in..."
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

# Create a small dummy video if it doesn't exist
if [ ! -f "test_video.mp4" ]; then
    echo "Creating dummy video..."
    ffmpeg -f lavfi -i color=c=blue:s=128x128:d=2 -c:v libx264 -t 2 test_video.mp4 -y
fi

echo -e "\nStep 3: Uploading 3 videos in parallel..."
for i in {1..3}; do
  echo "Uploading video $i..."
  curl -s -X POST "$API_URL/api/upload" \
    -H "Authorization: Bearer $TOKEN" \
    -F "video=@test_video.mp4" &
done

wait # Wait for all background uploads to finish

echo -e "\nStep 4: Checking status..."
for i in {1..5}; do
  echo "Poll $i:"
  curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/api/videos" | grep -oP '"status":"[^"]+"'
  sleep 2
done

echo -e "\nVerification complete!"

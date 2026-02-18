#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
TEST_VIDEO="test_video.mp4"
USER_EMAIL="test_$(date +%s)@example.com"
USER_PASS="password123"
DOWNLOAD_DIR="./downloads"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "🚀 ${GREEN}Starting Video Download Validation${NC}"

# 1. Create a valid test video file if it doesn't exist
if [ ! -f "$TEST_VIDEO" ]; then
    echo "Creating a valid test video file using FFmpeg..."
    if command -v ffmpeg &> /dev/null; then
        ffmpeg -y -f lavfi -i testsrc=duration=2:size=640x480:rate=30 -pix_fmt yuv420p "$TEST_VIDEO" -loglevel quiet
    else
        echo -e "⚠️  ${RED}FFmpeg not found!${NC} Please provide a valid test_video.mp4 file."
        exit 1
    fi
fi

# 2. Register test user
echo "Registering user..."
REGISTER_RES=$(curl -s -X POST "$API_URL/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$USER_EMAIL\", \"password\":\"$USER_PASS\", \"name\":\"Tester\"}")

# 3. Login to get JWT
echo "Logging in..."
LOGIN_RES=$(curl -s -X POST "$API_URL/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$USER_EMAIL\", \"password\":\"$USER_PASS\"}")

TOKEN=$(echo $LOGIN_RES | grep -oP '"token":"\K[^"]+')

if [ -z "$TOKEN" ]; then
    echo -e "❌ ${RED}Login failed${NC}"
    exit 1
fi

# 4. Upload video
echo "Uploading video..."
UPLOAD_RES=$(curl -s -X POST "$API_URL/api/upload" \
    -H "Authorization: Bearer $TOKEN" \
    -F "video=@$TEST_VIDEO")

VIDEO_ID=$(echo $UPLOAD_RES | grep -oP '"video_id":\K[0-9]+')

if [ -z "$VIDEO_ID" ]; then
    echo -e "❌ ${RED}Upload failed${NC}"
    echo $UPLOAD_RES
    exit 1
fi

echo -e "✅ Video uploaded with ID: ${GREEN}$VIDEO_ID${NC}"

# 5. Polling for completion
echo "Polling for processing completion..."
MAX_RETRIES=30
COUNT=0
ZIP_PATH=""

while [ $COUNT -lt $MAX_RETRIES ]; do
    STATUS_RES=$(curl -s -X GET "$API_URL/api/videos" \
        -H "Authorization: Bearer $TOKEN")
    
    # Simple grep to find the status of our video
    V_STATUS=$(echo $STATUS_RES | grep -oP "\"id\":$VIDEO_ID.*?\"status\":\"\K[^\"]+")
    
    if [ "$V_STATUS" == "COMPLETED" ]; then
        ZIP_PATH=$(echo $STATUS_RES | grep -oP "\"id\":$VIDEO_ID.*?\"zip_path\":\"\K[^\"]+")
        echo -e "✅ ${GREEN}Processing completed!${NC}"
        break
    elif [ "$V_STATUS" == "FAILED" ]; then
        echo -e "❌ ${RED}Processing failed${NC}"
        exit 1
    fi
    
    echo "Status: $V_STATUS... waiting (retry $((COUNT+1))/$MAX_RETRIES)"
    sleep 3
    COUNT=$((COUNT+1))
done

if [ -z "$ZIP_PATH" ]; then
    echo -e "❌ ${RED}Timeout reached or ZIP path not found${NC}"
    exit 1
fi

# 6. Download the ZIP via public endpoint
mkdir -p "$DOWNLOAD_DIR"
echo "Downloading ZIP: $ZIP_PATH..."
curl -s -O -J -L "$API_URL/download/$ZIP_PATH" --output-dir "$DOWNLOAD_DIR"

if [ -f "$DOWNLOAD_DIR/$ZIP_PATH" ]; then
    echo -e "🚀 ${GREEN}SUCCESS!${NC} File downloaded to $DOWNLOAD_DIR/$ZIP_PATH"
    ls -lh "$DOWNLOAD_DIR/$ZIP_PATH"
else
    echo -e "❌ ${RED}Download failed!${NC} File not found in $DOWNLOAD_DIR"
    exit 1
fi

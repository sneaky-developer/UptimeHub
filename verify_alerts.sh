#!/bin/bash
set -e

API_URL="http://localhost:8080/api/admin"

echo "Logging in..."
TOKEN=$(curl -s -X POST "$API_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@uptimehub.local","password":"admin123"}' | jq -r .token)

if [ "$TOKEN" == "null" ]; then
    echo "Login failed"
    exit 1
fi

echo "Got Token: ${TOKEN:0:10}..."

echo "Creating Webhook Channel..."
CREATE_RES=$(curl -s -X POST "$API_URL/alerts/channels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Webhook",
    "type": "webhook",
    "config": {"url": "https://httpbin.org/post"},
    "enabled": true
  }')

echo "$CREATE_RES" | jq .

CHANNEL_ID=$(echo "$CREATE_RES" | jq -r .id)

if [ "$CHANNEL_ID" == "null" ]; then
    echo "Creation failed"
    exit 1
fi

echo "Testing Channel $CHANNEL_ID..."
curl -s -X POST "$API_URL/alerts/channels/$CHANNEL_ID/test" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo "Listing Channels..."
curl -s -X GET "$API_URL/alerts/channels" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo "Done."

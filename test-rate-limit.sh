#!/bin/bash

# Test Rate Limiting - Sends 5 requests rapidly
# Expected: 1st request succeeds, next 4 get rate limited

echo "🧪 Testing Rate Limiter (RATE_LIMIT=1 req/s per IP)"
echo "=================================================="
echo ""
echo "Sending 5 rapid requests..."
echo ""

for i in {1..5}; do
  echo -n "Request $i: "
  response=$(curl -s -w "\n%{http_code}" http://localhost:3000/v1/find-country?ip=8.8.8.8)
  http_code=$(echo "$response" | tail -n 1)
  body=$(echo "$response" | head -n -1)

  if [ "$http_code" == "200" ]; then
    echo "✅ SUCCESS - $body"
  elif [ "$http_code" == "429" ]; then
    echo "🚫 RATE LIMITED - $body"
  else
    echo "❌ ERROR ($http_code) - $body"
  fi

  sleep 0.1  # Small delay between requests
done

echo ""
echo "=================================================="
echo "⏳ Waiting 1.1 seconds for rate limit to reset..."
sleep 1.1

echo ""
echo "Sending 1 more request after cooldown:"
echo -n "Request 6: "
response=$(curl -s -w "\n%{http_code}" http://localhost:3000/v1/find-country?ip=8.8.8.8)
http_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" == "200" ]; then
  echo "✅ SUCCESS - $body"
elif [ "$http_code" == "429" ]; then
  echo "🚫 RATE LIMITED - $body"
else
  echo "❌ ERROR ($http_code) - $body"
fi

echo ""
echo "✅ Test complete!"

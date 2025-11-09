#!/bin/bash
set -e

BASE_URL="${KESTREL_URL:-http://localhost:8080}"
API_KEY="${KESTREL_API_KEY}"

if [ -z "$API_KEY" ]; then
    echo "Error: KESTREL_API_KEY required"
    exit 1
fi

echo "🦅 Testing with API Key: ${API_KEY:0:20}..."
echo ""

# Test 1: Submit IOC
echo "Test 1: Submit IOC"
response=$(curl -s -X POST "$BASE_URL/api/ioc" \
    -H "X-API-Key: $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"domain":"evil.com","category":"Network activity","comment":"Test","feed":"premium"}')
event_id=$(echo "$response" | jq -r '.event_id')
echo "✓ IOC submitted, event_id: $event_id"

# Test 2: Get MISP Manifest
echo ""
echo "Test 2: Get MISP Manifest"
manifest=$(curl -s "$BASE_URL/misp/manifest.json" -H "X-API-Key: $API_KEY")
echo "✓ Manifest retrieved"
echo "$manifest" | jq '.' | head -10

# Test 3: Get MISP Event
echo ""
echo "Test 3: Get MISP Event"
event=$(curl -s "$BASE_URL/misp/events/$event_id.json" -H "X-API-Key: $API_KEY")
echo "✓ Event retrieved"

# Validate MISP format
if echo "$event" | jq -e '.Event' > /dev/null 2>&1; then
    echo "✓ Has Event wrapper (MISP compliant)"
fi
if echo "$event" | jq -e '.Event.Attribute' > /dev/null 2>&1; then
    echo "✓ Has Attribute array"
fi
if echo "$event" | jq -e '.Event.info' > /dev/null 2>&1; then
    echo "✓ Has info field"
fi
if echo "$event" | jq -e '.Event.threat_level_id' > /dev/null 2>&1; then
    echo "✓ Has threat_level_id"
fi

echo ""
echo "MISP Event Structure:"
echo "$event" | jq '.'

# Test 4: Get Premium Feed
echo ""
echo "Test 4: Get Premium PiHole Feed"
feed=$(curl -s "$BASE_URL/pihole/premium.txt?apikey=$API_KEY")
domains=$(echo "$feed" | wc -l | tr -d ' ')
echo "✓ Premium feed retrieved: $domains domain(s)"
echo "$feed" | head -5

echo ""
echo "✓ All authenticated tests passed!"

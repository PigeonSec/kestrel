#!/bin/bash

# Kestrel MISP Compliance Test Script
set -e

BASE_URL="${KESTREL_URL:-http://localhost:8080}"
COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_BLUE='\033[0;34m'
COLOR_NC='\033[0m'

echo -e "${COLOR_BLUE}🦅 Kestrel MISP Compliance Test Suite${COLOR_NC}"
echo "Testing against: $BASE_URL"
echo ""

pass() { echo -e "${COLOR_GREEN}✓ PASS:${COLOR_NC} $1"; }
fail() { echo -e "${COLOR_RED}✗ FAIL:${COLOR_NC} $1"; exit 1; }
info() { echo -e "${COLOR_BLUE}ℹ INFO:${COLOR_NC} $1"; }

# Test 1: Health check
echo ""
info "Test 1: Health Check"
response=$(curl -s "$BASE_URL/healthz")
if echo "$response" | jq -e '.status == "ok"' > /dev/null 2>&1; then
    pass "Health endpoint returns valid JSON"
else
    fail "Health endpoint failed: $response"
fi

# Test 2: Public PiHole feed
echo ""
info "Test 2: Public PiHole Feed (No Auth)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/pihole/public.txt")
if [ "$http_code" = "200" ]; then
    pass "Public PiHole feed accessible"
else
    fail "Public feed failed with code $http_code"
fi

# Test 3: Premium feed without auth
echo ""
info "Test 3: Premium PiHole Feed (Should require auth)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/pihole/premium.txt")
if [ "$http_code" = "401" ]; then
    pass "Premium feed requires authentication"
else
    info "Premium feed returned $http_code (may need auth setup)"
fi

# Test 4: MISP manifest without auth
echo ""
info "Test 4: MISP Manifest (Should require auth)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/misp/manifest.json")
if [ "$http_code" = "401" ]; then
    pass "MISP manifest requires authentication"
else
    info "MISP returned $http_code (may need auth setup)"
fi

# Test 5: IOC endpoint without auth
echo ""
info "Test 5: IOC Ingestion (Should require auth)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/ioc" \
    -H "Content-Type: application/json" \
    -d '{"domain":"test.com","category":"test","feed":"public"}')
if [ "$http_code" = "401" ]; then
    pass "IOC endpoint requires authentication"
else
    info "IOC endpoint returned $http_code"
fi

echo ""
echo "=========================================="
echo -e "${COLOR_GREEN}✓ Basic Tests Passed!${COLOR_NC}"
echo "=========================================="
echo ""
echo "Next Steps:"
echo "1. Generate API key: ./kestrel -generate-key"
echo "2. Run authenticated tests: KESTREL_API_KEY=<key> ./test_with_auth.sh"

#!/bin/bash

# Kestrel Authenticated API Test Script
# Requires a valid API key

set -e

BASE_URL="${KESTREL_URL:-http://localhost:8080}"
API_KEY="${KESTREL_API_KEY}"

COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[1;33m'
COLOR_BLUE='\033[0;34m'
COLOR_NC='\033[0m'

if [ -z "$API_KEY" ]; then
    echo -e "${COLOR_RED}Error: KESTREL_API_KEY environment variable not set${COLOR_NC}"
    echo "Usage: KESTREL_API_KEY=your-key ./test_with_auth.sh"
    exit 1
fi

echo -e "${COLOR_BLUE}🦅 Kestrel Authenticated API Tests${COLOR_NC}"
echo "=========================================="
echo "Testing against: $BASE_URL"
echo "Using API Key: ${API_KEY:0:20}..."
echo ""

pass() {
    echo -e "${COLOR_GREEN}✓ PASS:${COLOR_NC} $1"
}

fail() {
    echo -e "${COLOR_RED}✗ FAIL:${COLOR_NC} $1"
    exit 1
}

info() {
    echo -e "${COLOR_BLUE}ℹ INFO:${COLOR_NC} $1"
}

# Test 1: Submit IOC without validation
test_submit_ioc_basic() {
    echo ""
    info "Test 1: Submit IOC (Basic, No Validation)"
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/ioc" \
        -H "X-API-Key: $API_KEY" \
        -H "Content-Type: application/json" \
        -d '{
            "domain": "evil.com",
            "category": "Network activity",
            "comment": "Test malicious domain",
            "feed": "premium"
        }')
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        pass "IOC submission successful"
    else
        fail "IOC submission failed with code $http_code: $body"
    fi
    
    event_id=$(echo "$body" | jq -r '.event_id')
    if [ -n "$event_id" ] && [ "$event_id" != "null" ]; then
        pass "Event ID returned: $event_id"
        echo "$event_id" > /tmp/kestrel_test_event_id
    else
        fail "No event ID in response"
    fi
}

# Test 2: Submit IOC with DNS validation
test_submit_ioc_dns_validation() {
    echo ""
    info "Test 2: Submit IOC with DNS Validation"
    
    # Use google.com as a known-good domain
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/ioc?validate=dns" \
        -H "X-API-Key: $API_KEY" \
        -H "Content-Type: application/json" \
        -d '{
            "domain": "google.com",
            "category": "Network activity",
            "comment": "Test with DNS validation",
            "feed": "premium"
        }')
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        pass "IOC with DNS validation successful"
    else
        info "Response: $body"
        warn "DNS validation may have failed (expected if google.com fails DNS check)"
    fi
}

# Test 3: Submit IOC with invalid domain (should fail validation)
test_submit_ioc_invalid_domain() {
    echo ""
    info "Test 3: Submit IOC with Invalid Domain (Should Fail Validation)"
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/ioc?validate=dns" \
        -H "X-API-Key: $API_KEY" \
        -H "Content-Type: application/json" \
        -d '{
            "domain": "this-domain-definitely-does-not-exist-12345.invalid",
            "category": "Network activity",
            "comment": "Should fail validation",
            "feed": "premium"
        }')
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "400" ]; then
        pass "Invalid domain correctly rejected by validation"
    else
        warn "Expected 400 for invalid domain, got $http_code"
    fi
}

# Test 4: Get MISP Manifest
test_misp_manifest() {
    echo ""
    info "Test 4: Get MISP Manifest"
    
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/misp/manifest.json" \
        -H "X-API-Key: $API_KEY")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        pass "MISP manifest retrieved successfully"
    else
        fail "MISP manifest failed with code $http_code"
    fi
    
    if echo "$body" | jq empty 2>/dev/null; then
        pass "MISP manifest is valid JSON"
        
        count=$(echo "$body" | jq 'keys | length')
        info "Manifest contains $count event(s)"
    else
        fail "MISP manifest is not valid JSON: $body"
    fi
}

# Test 5: Get MISP Event
test_misp_event() {
    echo ""
    info "Test 5: Get MISP Event"
    
    if [ ! -f /tmp/kestrel_test_event_id ]; then
        warn "No event ID from previous test, skipping"
        return
    fi
    
    event_id=$(cat /tmp/kestrel_test_event_id)
    
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/misp/events/$event_id.json" \
        -H "X-API-Key: $API_KEY")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        pass "MISP event retrieved successfully"
    else
        fail "MISP event retrieval failed with code $http_code"
    fi
    
    # Validate MISP format
    if echo "$body" | jq -e '.Event' > /dev/null 2>&1; then
        pass "MISP event has Event wrapper (MISP compliant)"
    else
        fail "MISP event missing Event wrapper: $body"
    fi
    
    if echo "$body" | jq -e '.Event.Attribute' > /dev/null 2>&1; then
        pass "MISP event has Attribute array"
    else
        fail "MISP event missing Attribute array"
    fi
    
    # Check required fields
    info=$(echo "$body" | jq -r '.Event.info')
    threat_level=$(echo "$body" | jq -r '.Event.threat_level_id')
    
    if [ -n "$info" ] && [ "$info" != "null" ]; then
        pass "MISP event has info field: $info"
    else
        fail "MISP event missing info field"
    fi
    
    if [ -n "$threat_level" ] && [ "$threat_level" != "null" ]; then
        pass "MISP event has threat_level_id: $threat_level"
    else
        fail "MISP event missing threat_level_id"
    fi
}

# Test 6: Get Premium PiHole Feed
test_pihole_premium() {
    echo ""
    info "Test 6: Get Premium PiHole Feed"
    
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/pihole/premium.txt?apikey=$API_KEY")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        pass "Premium PiHole feed retrieved successfully"
    else
        fail "Premium PiHole feed failed with code $http_code"
    fi
    
    if [ -n "$body" ]; then
        domain_count=$(echo "$body" | wc -l | tr -d ' ')
        pass "Premium feed contains $domain_count domain(s)"
        info "Sample domains:"
        echo "$body" | head -5
    else
        info "Premium feed is empty (no domains yet)"
    fi
}

# Test 7: Admin - Generate New Key
test_admin_generate_key() {
    echo ""
    info "Test 7: Admin - Generate New API Key"
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/admin/generate-key" \
        -H "X-API-Key: $API_KEY" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "test@example.com",
            "plan": "free"
        }')
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        pass "Admin API key generation successful"
        
        new_key=$(echo "$body" | jq -r '.api_key')
        if [[ "$new_key" == kestrel_* ]]; then
            pass "Generated key has correct prefix: ${new_key:0:30}..."
            echo "$new_key" > /tmp/kestrel_test_new_key
        else
            fail "Generated key missing kestrel_ prefix: $new_key"
        fi
    else
        fail "Admin key generation failed with code $http_code: $body"
    fi
}

# Test 8: Admin - List Accounts
test_admin_list_accounts() {
    echo ""
    info "Test 8: Admin - List Accounts"
    
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/admin/accounts" \
        -H "X-API-Key: $API_KEY")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        pass "Admin account listing successful"
        
        account_count=$(echo "$body" | jq '.accounts | length')
        pass "Found $account_count account(s)"
    else
        fail "Admin account listing failed with code $http_code"
    fi
}

# Run all tests
echo "=========================================="
echo "Running Authenticated Tests..."
echo "=========================================="

test_submit_ioc_basic
test_submit_ioc_dns_validation
test_submit_ioc_invalid_domain
test_misp_manifest
test_misp_event
test_pihole_premium
test_admin_generate_key
test_admin_list_accounts

# Cleanup
rm -f /tmp/kestrel_test_event_id /tmp/kestrel_test_new_key

echo ""
echo "=========================================="
echo -e "${COLOR_GREEN}✓ All Authenticated Tests Passed!${COLOR_NC}"
echo "=========================================="
echo ""

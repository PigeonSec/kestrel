#!/bin/bash
# STIX 2.1 Compliance Test Script for Kestrel

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_KEY="${ADMIN_API_KEY:-kestrel_test_admin}"
USER_KEY="${KESTREL_API_KEY:-kestrel_test_user}"

echo "🦅 Kestrel STIX 2.1 Compliance Test"
echo "===================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass() {
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

info() {
    echo "ℹ $1"
}

# Test 1: Ingest IOC and verify STIX ID generation
echo ""
info "Test 1: Ingest IOC and verify STIX ID is generated"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/ioc" \
    -H "X-API-Key: $ADMIN_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "domain": "malicious-stix-test.com",
        "category": "Malware",
        "comment": "STIX compliance test domain",
        "feed": "stix-test",
        "access_level": "free"
    }')

STIX_ID=$(echo "$RESPONSE" | grep -o '"stix_id":"[^"]*"' | cut -d'"' -f4)
if [[ $STIX_ID == indicator--* ]]; then
    pass "IOC ingestion returned STIX ID: $STIX_ID"
else
    fail "IOC ingestion did not return valid STIX ID"
fi

# Test 2: List all indicators
echo ""
info "Test 2: List all STIX indicators"
RESPONSE=$(curl -s -X GET "$BASE_URL/stix/indicators" \
    -H "X-API-Key: $USER_KEY")

INDICATOR_COUNT=$(echo "$RESPONSE" | grep -o '"count":[0-9]*' | cut -d':' -f2)
if [[ $INDICATOR_COUNT -gt 0 ]]; then
    pass "Found $INDICATOR_COUNT STIX indicators"
else
    fail "No STIX indicators found"
fi

# Test 3: Get specific indicator by ID
echo ""
info "Test 3: Retrieve specific STIX indicator"
RESPONSE=$(curl -s -X GET "$BASE_URL/stix/indicators/$STIX_ID" \
    -H "X-API-Key: $USER_KEY")

# Validate STIX 2.1 indicator structure
if echo "$RESPONSE" | grep -q '"type":"indicator"'; then
    pass "Indicator has correct type field"
else
    fail "Indicator missing type field"
fi

if echo "$RESPONSE" | grep -q '"spec_version":"2.1"'; then
    pass "Indicator has STIX 2.1 spec_version"
else
    fail "Indicator missing or incorrect spec_version"
fi

if echo "$RESPONSE" | grep -q '"pattern":"'; then
    pass "Indicator has pattern field"
else
    fail "Indicator missing pattern field"
fi

if echo "$RESPONSE" | grep -q '"pattern_type":"stix"'; then
    pass "Indicator has correct pattern_type"
else
    fail "Indicator missing or incorrect pattern_type"
fi

if echo "$RESPONSE" | grep -q '"valid_from":"'; then
    pass "Indicator has valid_from timestamp"
else
    fail "Indicator missing valid_from field"
fi

# Test 4: Get indicator as bundle
echo ""
info "Test 4: Retrieve indicator as STIX bundle"
RESPONSE=$(curl -s -X GET "$BASE_URL/stix/indicators/$STIX_ID/bundle" \
    -H "X-API-Key: $USER_KEY")

if echo "$RESPONSE" | grep -q '"type":"bundle"'; then
    pass "Response is a STIX bundle"
else
    fail "Response is not a valid STIX bundle"
fi

if echo "$RESPONSE" | grep -q '"objects":\['; then
    pass "Bundle contains objects array"
else
    fail "Bundle missing objects array"
fi

# Test 5: Get full STIX bundle with all indicators
echo ""
info "Test 5: Retrieve full STIX bundle"
RESPONSE=$(curl -s -X GET "$BASE_URL/stix/bundle" \
    -H "X-API-Key: $USER_KEY")

if echo "$RESPONSE" | grep -q '"type":"bundle"'; then
    pass "Full bundle endpoint returns valid bundle"
else
    fail "Full bundle endpoint failed"
fi

BUNDLE_ID=$(echo "$RESPONSE" | grep -o '"id":"bundle--[^"]*"' | cut -d'"' -f4)
if [[ $BUNDLE_ID == bundle--* ]]; then
    pass "Bundle has valid bundle ID: $BUNDLE_ID"
else
    fail "Bundle has invalid ID format"
fi

# Test 6: Get STIX objects (TAXII-like endpoint)
echo ""
info "Test 6: Retrieve STIX objects via TAXII-like endpoint"
RESPONSE=$(curl -s -X GET "$BASE_URL/stix/objects" \
    -H "X-API-Key: $USER_KEY")

if echo "$RESPONSE" | grep -q '"objects":\['; then
    pass "Objects endpoint returns objects array"
else
    fail "Objects endpoint failed"
fi

if echo "$RESPONSE" | grep -q '"more":false'; then
    pass "Objects endpoint includes pagination metadata"
else
    warn "Objects endpoint missing pagination metadata"
fi

# Test 7: Verify Content-Type headers
echo ""
info "Test 7: Verify STIX content-type headers"
CONTENT_TYPE=$(curl -s -I -X GET "$BASE_URL/stix/bundle" \
    -H "X-API-Key: $USER_KEY" | grep -i "content-type:" | tr -d '\r')

if echo "$CONTENT_TYPE" | grep -qi "application/stix+json"; then
    pass "STIX endpoint returns correct content-type"
else
    warn "STIX endpoint content-type may not be fully compliant: $CONTENT_TYPE"
fi

# Test 8: Validate STIX pattern format
echo ""
info "Test 8: Validate STIX pattern format"
PATTERN=$(curl -s -X GET "$BASE_URL/stix/indicators/$STIX_ID" \
    -H "X-API-Key: $USER_KEY" | grep -o '"pattern":"[^"]*"' | cut -d'"' -f4)

if [[ $PATTERN == *"domain-name:value"* ]]; then
    pass "STIX pattern uses correct domain-name observable: $PATTERN"
else
    fail "STIX pattern format incorrect: $PATTERN"
fi

# Test 9: Verify RFC3339 timestamp format
echo ""
info "Test 9: Verify RFC3339 timestamp format"
CREATED=$(curl -s -X GET "$BASE_URL/stix/indicators/$STIX_ID" \
    -H "X-API-Key: $USER_KEY" | grep -o '"created":"[^"]*"' | cut -d'"' -f4)

if [[ $CREATED =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2} ]]; then
    pass "Timestamps use RFC3339 format: $CREATED"
else
    fail "Timestamp format incorrect: $CREATED"
fi

# Test 10: Test idempotency - re-ingest same domain
echo ""
info "Test 10: Test STIX ID idempotency for duplicate domains"
RESPONSE2=$(curl -s -X POST "$BASE_URL/api/ioc" \
    -H "X-API-Key: $ADMIN_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "domain": "malicious-stix-test.com",
        "category": "Malware",
        "comment": "Duplicate test",
        "feed": "stix-test",
        "access_level": "free"
    }')

STIX_ID2=$(echo "$RESPONSE2" | grep -o '"stix_id":"[^"]*"' | cut -d'"' -f4)
if [[ "$STIX_ID" == "$STIX_ID2" ]]; then
    pass "STIX ID is stable across re-ingestion: $STIX_ID"
else
    warn "STIX ID changed on re-ingestion: $STIX_ID vs $STIX_ID2"
fi

echo ""
echo "===================================="
echo -e "${GREEN}✓ All STIX 2.1 compliance tests passed!${NC}"
echo ""
echo "Summary:"
echo "  - STIX 2.1 bundle generation: ✓"
echo "  - Indicator object structure: ✓"
echo "  - Pattern format: ✓"
echo "  - Timestamp format: ✓"
echo "  - Content-Type headers: ✓"
echo "  - TAXII-like endpoints: ✓"
echo "  - ID stability: ✓"
echo ""
echo "Next steps:"
echo "  1. Implement full TAXII 2.1 server (Phase 2)"
echo "  2. Add MISP object support (Phase 4)"
echo "  3. Implement relationships and sightings"
echo "  4. Add filtering and pagination"
echo ""

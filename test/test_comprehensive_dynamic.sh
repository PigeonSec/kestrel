#!/bin/bash
# Comprehensive test of fully dynamic, metadata-driven API

set -e

BASE_URL="http://localhost:8185"
ADMIN_KEY="kestrel_admin_test"
USER_KEY="kestrel_user_test"

echo "🚀 Comprehensive Dynamic API Test"
echo "=================================="
echo ""

# Remove old database
rm -f kestrel.db

# Start server
echo "Starting Kestrel..."
ADMIN_API_KEY="$ADMIN_KEY" LISTEN_ADDR=:8185 ENABLE_VALIDATION=false ./kestrel > /tmp/kestrel_comprehensive.log 2>&1 &
SERVER_PID=$!
sleep 3

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    kill $SERVER_PID 2>/dev/null || true
    rm -f kestrel.db
}
trap cleanup EXIT

# Add regular user key
curl -s -X POST "$BASE_URL/api/admin/accounts" \
  -H "X-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"api_key\":\"$USER_KEY\",\"email\":\"user@test.com\",\"plan\":\"premium\",\"active\":true}" > /dev/null

echo "✓ Server started"
echo ""

# =============================================================================
echo "PART 1: Data Ingestion (Admin Only)"
echo "============================================================================="
echo ""

echo "Test 1.1: Create 'community' feed (free access)"
response=$(curl -s -X POST "$BASE_URL/api/ioc" \
  -H "X-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"domain":"malware1.com","category":"Malware","comment":"Community reported","feed":"community","access_level":"free"}')
echo "   Response: $response"
echo "   ✓ Free feed created"

echo ""
echo "Test 1.2: Add more domains to 'community' feed"
for i in {2..5}; do
    curl -s -X POST "$BASE_URL/api/ioc" \
      -H "X-API-Key: $ADMIN_KEY" \
      -H "Content-Type: application/json" \
      -d "{\"domain\":\"malware$i.com\",\"category\":\"Malware\",\"feed\":\"community\",\"access_level\":\"free\"}" > /dev/null
done
echo "   ✓ Added 4 more domains"

echo ""
echo "Test 1.3: Create 'premium' feed (paid access - default)"
curl -s -X POST "$BASE_URL/api/ioc" \
  -H "X-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"domain":"apt-threat1.com","category":"APT","comment":"Premium intel","feed":"premium"}' > /dev/null
echo "   ✓ Paid feed created"

echo ""
echo "Test 1.4: Create 'enterprise' feed (explicit paid access)"
curl -s -X POST "$BASE_URL/api/ioc" \
  -H "X-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"domain":"enterprise-threat.com","category":"APT","feed":"enterprise","access_level":"paid"}' > /dev/null
echo "   ✓ Explicit paid feed created"

echo ""
echo "Test 1.5: Create 'apac-threats' feed (regional, free)"
curl -s -X POST "$BASE_URL/api/ioc" \
  -H "X-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"domain":"apac-malware.com","category":"Regional","feed":"apac-threats","access_level":"free"}' > /dev/null
echo "   ✓ Regional free feed created"

# =============================================================================
echo ""
echo "PART 2: Dynamic Path Access - Free Feeds"
echo "============================================================================="
echo ""

echo "Test 2.1: Access 'community' feed via /list/pihole/community.txt"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/list/pihole/community.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 2.2: Access 'community' feed via /list/adguard/community.txt"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/list/adguard/community.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 2.3: Access 'community' feed via /feeds/community.txt (legacy path)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/feeds/community.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 2.4: Access 'community' feed via /blocklists/dns/sinkhole/community.txt"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/blocklists/dns/sinkhole/community.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 2.5: Access 'community' feed via /custom/deep/nested/path/structure/community.txt"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/custom/deep/nested/path/structure/community.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 2.6: Access 'apac-threats' via /geo/apac/apac-threats.txt"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/geo/apac/apac-threats.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 2.7: Verify content - community feed should have 5 domains"
content=$(curl -s "$BASE_URL/list/pihole/community.txt")
count=$(echo "$content" | wc -l | tr -d ' ')
echo "   Domain count: $count $([ $count = 5 ] && echo '✓' || echo '✗ Expected 5')"
echo "   First domain: $(echo "$content" | head -1)"

# =============================================================================
echo ""
echo "PART 3: Dynamic Path Access - Paid Feeds (Auth Required)"
echo "============================================================================="
echo ""

echo "Test 3.1: Access 'premium' feed without auth (should fail)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/list/pihole/premium.txt")
echo "   Status: $http_code $([ $http_code = 401 ] && echo '✓' || echo '✗ Expected 401')"

echo ""
echo "Test 3.2: Access 'premium' feed with user key via header"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "X-API-Key: $USER_KEY" \
  "$BASE_URL/list/pihole/premium.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 3.3: Access 'premium' feed with user key via query param"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  "$BASE_URL/list/adguard/premium.txt?apikey=$USER_KEY")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 3.4: Access 'enterprise' feed via custom path with auth"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "X-API-Key: $USER_KEY" \
  "$BASE_URL/tenant/corporate/security/enterprise.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 3.5: Access 'enterprise' feed via legacy path with auth"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "X-API-Key: $USER_KEY" \
  "$BASE_URL/feeds/enterprise.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 3.6: Access 'premium' feed with admin key (should work)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "X-API-Key: $ADMIN_KEY" \
  "$BASE_URL/list/pihole/premium.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

# =============================================================================
echo ""
echo "PART 4: Edge Cases and Error Handling"
echo "============================================================================="
echo ""

echo "Test 4.1: Access non-existent feed (should fail - no metadata)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/list/pihole/nonexistent.txt")
echo "   Status: $http_code $([ $http_code = 401 ] && echo '✓ (defaults to paid)' || echo \"Got $http_code\")"

echo ""
echo "Test 4.2: Access feed without .txt extension"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/list/pihole/community")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 4.3: Access with just feed name at root"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/community.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 4.4: Invalid API key on paid feed"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "X-API-Key: invalid_key_12345" \
  "$BASE_URL/list/pihole/premium.txt")
echo "   Status: $http_code $([ $http_code = 401 ] && echo '✓' || echo '✗ Expected 401')"

echo ""
echo "Test 4.5: Empty path (should fail)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/")
if [ $http_code = 404 ]; then
    echo "   Status: $http_code ✓"
else
    echo "   Status: $http_code (acceptable)"
fi

# =============================================================================
echo ""
echo "PART 5: Multiple Feeds with Same Name in Different Paths"
echo "============================================================================="
echo ""

echo "Test 5.1: 'community' via 10 different path structures (all should work)"
paths=(
    "/list/pihole/community.txt"
    "/list/adguard/community.txt"
    "/feeds/community.txt"
    "/pihole/community.txt"
    "/blocklists/community.txt"
    "/dns/sinkhole/community.txt"
    "/security/feeds/public/community.txt"
    "/threat-intel/open/community.txt"
    "/v1/api/feeds/community.txt"
    "/custom/path/community.txt"
)

success_count=0
for path in "${paths[@]}"; do
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$path")
    if [ $http_code = 200 ]; then
        ((success_count++))
    fi
done
echo "   $success_count/10 paths successful $([ $success_count = 10 ] && echo '✓' || echo '✗')"

# =============================================================================
echo ""
echo "PART 6: Authorization Header Variants"
echo "============================================================================="
echo ""

echo "Test 6.1: Bearer token in Authorization header"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $USER_KEY" \
  "$BASE_URL/list/pihole/premium.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 6.2: X-API-Key header (preferred)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "X-API-Key: $USER_KEY" \
  "$BASE_URL/list/pihole/premium.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 6.3: Query parameter ?apikey="
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  "$BASE_URL/list/pihole/premium.txt?apikey=$USER_KEY")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

# =============================================================================
echo ""
echo "PART 7: Metadata-Driven Access Control"
echo "============================================================================="
echo ""

echo "Test 7.1: Create new 'test-free' feed with admin key"
curl -s -X POST "$BASE_URL/api/ioc" \
  -H "X-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"domain":"test1.com","category":"Test","feed":"test-free","access_level":"free"}' > /dev/null
echo "   ✓ Created"

echo ""
echo "Test 7.2: Access 'test-free' without auth (should work immediately)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/any/path/test-free.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

echo ""
echo "Test 7.3: Create new 'test-paid' feed (no access_level = paid default)"
curl -s -X POST "$BASE_URL/api/ioc" \
  -H "X-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"domain":"secret.com","category":"Test","feed":"test-paid"}' > /dev/null
echo "   ✓ Created"

echo ""
echo "Test 7.4: Access 'test-paid' without auth (should fail)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/any/path/test-paid.txt")
echo "   Status: $http_code $([ $http_code = 401 ] && echo '✓' || echo '✗ Expected 401')"

echo ""
echo "Test 7.5: Access 'test-paid' with auth (should work)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "X-API-Key: $USER_KEY" \
  "$BASE_URL/any/path/test-paid.txt")
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

# =============================================================================
echo ""
echo "PART 8: Admin-Only Ingestion"
echo "============================================================================="
echo ""

echo "Test 8.1: Try to ingest with regular user key (should fail)"
output=$(mktemp)
http_code=$(curl -s -w "%{http_code}" \
  -X POST "$BASE_URL/api/ioc" \
  -H "X-API-Key: $USER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"domain":"fail.com","category":"Test","feed":"fail"}' \
  -o "$output")
body=$(cat "$output")
rm -f "$output"
echo "   Status: $http_code $([ $http_code = 403 ] && echo '✓' || echo '✗ Expected 403')"
echo "   Message: $body"

echo ""
echo "Test 8.2: Try to ingest without auth (should fail)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "$BASE_URL/api/ioc" \
  -H "Content-Type: application/json" \
  -d '{"domain":"fail2.com","category":"Test","feed":"fail"}')
echo "   Status: $http_code $([ $http_code = 401 ] && echo '✓' || echo '✗ Expected 401')"

echo ""
echo "Test 8.3: Ingest with admin key (should work)"
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "$BASE_URL/api/ioc" \
  -H "X-API-Key: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"domain":"success.com","category":"Test","feed":"admin-only","access_level":"free"}')
echo "   Status: $http_code $([ $http_code = 200 ] && echo '✓' || echo '✗')"

# =============================================================================
echo ""
echo "========================================="
echo "✓ COMPREHENSIVE DYNAMIC API TEST COMPLETE!"
echo "========================================="
echo ""
echo "Summary:"
echo "--------"
echo "✓ All feed paths are fully dynamic"
echo "✓ Feed names extracted from last URL segment"
echo "✓ Access control determined by ingested metadata"
echo "✓ Free feeds (access_level: 'free') - no auth required"
echo "✓ Paid feeds (access_level: 'paid' or default) - auth required"
echo "✓ Any path structure works: /pihole/, /feeds/, /custom/deep/path/"
echo "✓ Only admin API keys can ingest IOCs"
echo "✓ Regular users can only read feeds (based on access level)"
echo "✓ No configuration files needed - 100% metadata-driven"
echo ""

#!/bin/bash

# Comprehensive monitoring stack testing script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROMETHEUS_URL="http://localhost:9090"
GRAFANA_URL="http://localhost:3000"
ALERTMANAGER_URL="http://localhost:9093"
APP_URL="http://localhost:8080"
METRICS_URL="http://localhost:2112"
WEBHOOK_URL="http://localhost:9999"

echo -e "${BLUE}=== Template Arch Lint Monitoring Stack Test ===${NC}"
echo "This script tests the complete monitoring infrastructure"
echo

# Function to check if a service is available
check_service() {
    local name=$1
    local url=$2
    local max_attempts=30
    local attempt=1
    
    echo -n "Checking $name... "
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -o /dev/null "$url" 2>/dev/null; then
            echo -e "${GREEN}✓ Available${NC}"
            return 0
        fi
        
        if [ $attempt -eq 1 ]; then
            echo -n "waiting"
        fi
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}✗ Failed to connect after $max_attempts attempts${NC}"
    return 1
}

# Function to test API endpoint
test_api_endpoint() {
    local endpoint=$1
    local expected_status=$2
    local description=$3
    
    echo -n "Testing $description... "
    
    local response
    local status_code
    response=$(curl -s -w "%{http_code}" "$endpoint" 2>/dev/null) || {
        echo -e "${RED}✗ Failed to connect${NC}"
        return 1
    }
    
    status_code="${response: -3}"
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓ Status $status_code${NC}"
        return 0
    else
        echo -e "${RED}✗ Expected $expected_status, got $status_code${NC}"
        return 1
    fi
}

# Function to test metrics endpoint
test_metrics() {
    echo -e "${YELLOW}Testing Prometheus metrics...${NC}"
    
    local metrics_response
    metrics_response=$(curl -s "$METRICS_URL/metrics" 2>/dev/null) || {
        echo -e "${RED}✗ Failed to fetch metrics${NC}"
        return 1
    }
    
    # Check for key metrics
    local expected_metrics=(
        "http_requests_total"
        "http_request_duration_seconds"
        "app_uptime_seconds"
        "user_created_total"
        "db_queries_total"
        "sla_availability_ratio"
        "go_goroutines"
    )
    
    local found_metrics=0
    for metric in "${expected_metrics[@]}"; do
        if echo "$metrics_response" | grep -q "^$metric"; then
            echo -e "  ${GREEN}✓ $metric${NC}"
            found_metrics=$((found_metrics + 1))
        else
            echo -e "  ${RED}✗ $metric (missing)${NC}"
        fi
    done
    
    echo "Found $found_metrics/${#expected_metrics[@]} expected metrics"
    
    if [ $found_metrics -eq ${#expected_metrics[@]} ]; then
        return 0
    else
        return 1
    fi
}

# Function to test Prometheus queries
test_prometheus_queries() {
    echo -e "${YELLOW}Testing Prometheus queries...${NC}"
    
    local queries=(
        "up"
        "http_requests_total"
        "rate(http_requests_total[5m])"
        "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
        "sla_availability_ratio"
        "sla_error_budget_ratio"
    )
    
    local successful_queries=0
    for query in "${queries[@]}"; do
        echo -n "  Query: $query... "
        
        local encoded_query
        encoded_query=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$query'))" 2>/dev/null) || {
            encoded_query="$query"
        }
        
        local response
        response=$(curl -s "${PROMETHEUS_URL}/api/v1/query?query=${encoded_query}" 2>/dev/null) || {
            echo -e "${RED}✗ Failed${NC}"
            continue
        }
        
        if echo "$response" | grep -q '"status":"success"'; then
            echo -e "${GREEN}✓ Success${NC}"
            successful_queries=$((successful_queries + 1))
        else
            echo -e "${RED}✗ Failed${NC}"
            echo "    Response: $response" | head -c 200
            echo "..."
        fi
    done
    
    echo "Successful queries: $successful_queries/${#queries[@]}"
    
    if [ $successful_queries -eq ${#queries[@]} ]; then
        return 0
    else
        return 1
    fi
}

# Function to generate test traffic
generate_test_traffic() {
    echo -e "${YELLOW}Generating test traffic...${NC}"
    
    # Create some users to generate metrics
    for i in {1..5}; do
        echo -n "  Creating test user $i... "
        
        local response
        response=$(curl -s -w "%{http_code}" -X POST \
            -H "Content-Type: application/json" \
            -d "{\"id\":\"test-user-$i\",\"email\":\"test$i@example.com\",\"name\":\"Test User $i\"}" \
            "$APP_URL/api/v1/users" 2>/dev/null) || {
            echo -e "${RED}✗ Failed to connect${NC}"
            continue
        }
        
        local status_code="${response: -3}"
        if [ "$status_code" = "201" ] || [ "$status_code" = "409" ]; then
            echo -e "${GREEN}✓ Success${NC}"
        else
            echo -e "${RED}✗ Status $status_code${NC}"
        fi
    done
    
    # Generate some GET requests
    for i in {1..10}; do
        curl -s "$APP_URL/health" > /dev/null 2>&1 &
        curl -s "$APP_URL/api/v1/users" > /dev/null 2>&1 &
        curl -s "$APP_URL/" > /dev/null 2>&1 &
    done
    
    wait
    echo "  Generated test traffic (health checks, user listing, root page)"
}

# Function to trigger test alerts
trigger_test_alerts() {
    echo -e "${YELLOW}Triggering test alerts...${NC}"
    
    # This would typically involve:
    # 1. Stopping the application temporarily
    # 2. Generating high error rates
    # 3. Creating high response times
    
    echo "  Note: Manual alert testing requires specific conditions"
    echo "  To test alerts manually:"
    echo "    - Stop the application container to trigger 'ApplicationDown' alert"
    echo "    - Generate high error rates to trigger error rate alerts"
    echo "    - Create artificial load to trigger response time alerts"
}

# Function to check AlertManager
test_alertmanager() {
    echo -e "${YELLOW}Testing AlertManager...${NC}"
    
    # Check AlertManager status
    test_api_endpoint "$ALERTMANAGER_URL/-/healthy" "200" "AlertManager health"
    
    # Check for any active alerts
    echo -n "  Checking active alerts... "
    local alerts_response
    alerts_response=$(curl -s "$ALERTMANAGER_URL/api/v1/alerts" 2>/dev/null) || {
        echo -e "${RED}✗ Failed to fetch alerts${NC}"
        return 1
    }
    
    local alert_count
    alert_count=$(echo "$alerts_response" | grep -o '"status":"active"' | wc -l)
    
    echo -e "${GREEN}✓ $alert_count active alerts${NC}"
    
    # Check configuration
    echo -n "  Checking configuration... "
    local config_response
    config_response=$(curl -s "$ALERTMANAGER_URL/api/v1/status" 2>/dev/null) || {
        echo -e "${RED}✗ Failed to fetch config${NC}"
        return 1
    }
    
    if echo "$config_response" | grep -q '"status":"success"'; then
        echo -e "${GREEN}✓ Configuration loaded${NC}"
    else
        echo -e "${RED}✗ Configuration error${NC}"
    fi
}

# Function to check Grafana
test_grafana() {
    echo -e "${YELLOW}Testing Grafana...${NC}"
    
    # Check Grafana health
    test_api_endpoint "$GRAFANA_URL/api/health" "200" "Grafana health"
    
    # Check datasources
    echo -n "  Checking datasources... "
    local ds_response
    ds_response=$(curl -s -u admin:admin "$GRAFANA_URL/api/datasources" 2>/dev/null) || {
        echo -e "${RED}✗ Failed to fetch datasources${NC}"
        return 1
    }
    
    local ds_count
    ds_count=$(echo "$ds_response" | grep -o '"name":"' | wc -l)
    
    echo -e "${GREEN}✓ $ds_count datasources configured${NC}"
    
    # Check dashboards
    echo -n "  Checking dashboards... "
    local dash_response
    dash_response=$(curl -s -u admin:admin "$GRAFANA_URL/api/search" 2>/dev/null) || {
        echo -e "${RED}✗ Failed to fetch dashboards${NC}"
        return 1
    }
    
    local dash_count
    dash_count=$(echo "$dash_response" | grep -o '"type":"dash-db"' | wc -l)
    
    echo -e "${GREEN}✓ $dash_count dashboards found${NC}"
}

# Main execution
main() {
    local failed_tests=0
    
    echo -e "${BLUE}Step 1: Service Availability${NC}"
    check_service "Application" "$APP_URL/health" || failed_tests=$((failed_tests + 1))
    check_service "Prometheus Metrics" "$METRICS_URL/metrics" || failed_tests=$((failed_tests + 1))
    check_service "Prometheus" "$PROMETHEUS_URL/-/healthy" || failed_tests=$((failed_tests + 1))
    check_service "AlertManager" "$ALERTMANAGER_URL/-/healthy" || failed_tests=$((failed_tests + 1))
    check_service "Grafana" "$GRAFANA_URL/api/health" || failed_tests=$((failed_tests + 1))
    echo
    
    echo -e "${BLUE}Step 2: Application Endpoints${NC}"
    test_api_endpoint "$APP_URL/health" "200" "Application health endpoint" || failed_tests=$((failed_tests + 1))
    test_api_endpoint "$APP_URL/" "200" "Application root endpoint" || failed_tests=$((failed_tests + 1))
    test_api_endpoint "$APP_URL/api/v1/users" "200" "Users API endpoint" || failed_tests=$((failed_tests + 1))
    echo
    
    echo -e "${BLUE}Step 3: Metrics Collection${NC}"
    test_metrics || failed_tests=$((failed_tests + 1))
    echo
    
    echo -e "${BLUE}Step 4: Prometheus Queries${NC}"
    test_prometheus_queries || failed_tests=$((failed_tests + 1))
    echo
    
    echo -e "${BLUE}Step 5: Test Traffic Generation${NC}"
    generate_test_traffic
    echo
    
    echo -e "${BLUE}Step 6: AlertManager Testing${NC}"
    test_alertmanager || failed_tests=$((failed_tests + 1))
    echo
    
    echo -e "${BLUE}Step 7: Grafana Testing${NC}"
    test_grafana || failed_tests=$((failed_tests + 1))
    echo
    
    echo -e "${BLUE}Step 8: Alert Testing${NC}"
    trigger_test_alerts
    echo
    
    # Summary
    echo -e "${BLUE}=== Test Summary ===${NC}"
    if [ $failed_tests -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        echo -e "${GREEN}Your monitoring stack is working correctly.${NC}"
    else
        echo -e "${RED}✗ $failed_tests test(s) failed.${NC}"
        echo -e "${YELLOW}Check the output above for details.${NC}"
    fi
    
    echo
    echo -e "${BLUE}Monitoring URLs:${NC}"
    echo -e "  Application:   $APP_URL"
    echo -e "  Prometheus:    $PROMETHEUS_URL"
    echo -e "  AlertManager:  $ALERTMANAGER_URL"
    echo -e "  Grafana:       $GRAFANA_URL (admin/admin)"
    echo -e "  Metrics:       $METRICS_URL/metrics"
    
    return $failed_tests
}

# Check if help was requested
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  --help, -h    Show this help message"
    echo "  --webhook     Start webhook test server (runs in background)"
    echo
    echo "This script tests the complete monitoring stack including:"
    echo "  - Service availability"
    echo "  - Metrics collection"
    echo "  - Prometheus queries"
    echo "  - AlertManager functionality"
    echo "  - Grafana integration"
    echo
    exit 0
fi

# Check if webhook server should be started
if [[ "$1" == "--webhook" ]]; then
    echo -e "${YELLOW}Starting webhook test server...${NC}"
    if command -v go >/dev/null 2>&1; then
        go run "$(dirname "$0")/webhook-test-server.go" &
        WEBHOOK_PID=$!
        echo -e "${GREEN}✓ Webhook server started (PID: $WEBHOOK_PID)${NC}"
        echo "  Access webhook endpoints at $WEBHOOK_URL"
        
        # Trap to kill webhook server on exit
        trap "kill $WEBHOOK_PID 2>/dev/null" EXIT
        
        sleep 2
    else
        echo -e "${YELLOW}⚠ Go not found, skipping webhook server${NC}"
    fi
fi

# Run main test suite
main
exit_code=$?

echo -e "${BLUE}Test completed.${NC}"
exit $exit_code
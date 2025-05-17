#!/bin/bash

# This script runs all tests for the Meals application

# Set test environment variables
export APP_ENV=test
export TEST_DATABASE_HOST=${TEST_DATABASE_HOST:-localhost}
export TEST_DATABASE_PORT=${TEST_DATABASE_PORT:-5432}
export TEST_DATABASE_USER=${TEST_DATABASE_USER:-postgres}
export TEST_DATABASE_PASSWORD=${TEST_DATABASE_PASSWORD:-postgres}
export TEST_DATABASE_NAME=${TEST_DATABASE_NAME:-meals_test}

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running tests for Meals Application${NC}"
echo "====================================="
echo "Test environment: APP_ENV=$APP_ENV"
echo "Test database: $TEST_DATABASE_NAME"
echo "====================================="

# Function to run tests with proper output
run_tests() {
    test_path=$1
    test_name=$2
    
    echo -e "\n${YELLOW}Running $test_name tests...${NC}"
    go test -v $test_path
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $test_name tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ $test_name tests failed${NC}"
        return 1
    fi
}

# Create directory for test reports if it doesn't exist
mkdir -p ./test-reports

# Run all tests or specific test packages based on arguments
if [ $# -eq 0 ]; then
    # Run all tests
    echo -e "\n${YELLOW}Running all tests...${NC}"
    result=0
    
    run_tests "./tests/models/..." "Models" || result=1
    run_tests "./tests/auth/..." "Authentication" || result=1
    run_tests "./tests/handlers/..." "Error Handling" || result=1
    run_tests "./tests/store/..." "Database & Transactions" || result=1
    run_tests "./tests/routes/..." "API Endpoints" || result=1
    
    if [ $result -eq 0 ]; then
        echo -e "\n${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "\n${RED}Some tests failed!${NC}"
        exit 1
    fi
else
    # Run specific test package
    for pkg in "$@"; do
        case $pkg in
            "models")
                run_tests "./tests/models/..." "Models"
                ;;
            "auth")
                run_tests "./tests/auth/..." "Authentication"
                ;;
            "handlers")
                run_tests "./tests/handlers/..." "Error Handling"
                ;;
            "store")
                run_tests "./tests/store/..." "Database & Transactions"
                ;;
            "routes")
                run_tests "./tests/routes/..." "API Endpoints"
                ;;
            *)
                echo -e "${RED}Unknown test package: $pkg${NC}"
                echo "Available packages: models, auth, handlers, store, routes"
                exit 1
                ;;
        esac
    done
fi

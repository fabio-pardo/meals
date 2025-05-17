# Test Suite for Meals App

This directory contains a comprehensive test suite for the Meals delivery application. The test suite covers various components including models, authentication, API endpoints, and utilities.

## Test Organization

The tests are organized in a modular structure:

- `tests/config.go` - Test configuration and utility functions
- `tests/main_test.go` - Main test entry point
- `tests/models/` - Tests for the data models and validation logic
- `tests/auth/` - Tests for the authentication and authorization system
- `tests/handlers/` - Tests for the error handling system
- `tests/store/` - Tests for the database and transaction utilities
- `tests/routes/` - Integration tests for API endpoints

## Setup Requirements

Before running tests, ensure you have:

1. A PostgreSQL test database configured
2. The environment variables for test database connection are set:
   - Set `APP_ENV=test` or create a `.env.test` file with your test database configuration

## Running Tests

To run all tests:

```bash
go test ./tests/...
```

To run specific test packages:

```bash
go test ./tests/models  # Run only model tests
go test ./tests/auth    # Run only authentication tests
```

To run a specific test:

```bash
go test ./tests/models -run TestOrderValidation
```

With verbose output:

```bash
go test ./tests/... -v
```

## Testing Patterns

The test suite follows these patterns:

1. **Database Setup**: Each test uses a clean database environment
2. **Transaction Tests**: Database operations are tested with transactions to ensure data integrity
3. **Role-Based Auth**: Authentication middleware is tested for all user roles
4. **Error Handling**: The error handling system is tested for various error types
5. **API Integration**: API endpoints are tested for correct behavior with authenticated and unauthenticated users

## Adding New Tests

When adding new tests:

1. Follow the established pattern for your test category
2. Use the utility functions in `tests/config.go` for database setup
3. Clear the database between tests using `SetupTest()`
4. Add test helpers as needed to reduce duplication

## Test Configuration

Tests use a separate database configuration from the production app. The test configuration is loaded from environment variables or a `.env.test` file.

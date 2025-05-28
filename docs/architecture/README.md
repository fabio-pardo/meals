# Architecture Overview

## System Components

### Web Server
- **Framework**: Gin-based REST API
- **Port**: Configurable (default: 8080)
- **Middleware**: Request ID, logging, recovery, authentication

### Authentication & Authorization
- **Primary**: OAuth2 (Google) with session management
- **Session Storage**: Cookie-based sessions with GORM backend
- **Authorization**: Role-based access control (Customer/Driver/Admin)
- **Middleware**: `auth.RequireAuth()`, `auth.RequireRole()`

### Data Storage
- **Primary Database**: PostgreSQL with GORM ORM
- **Cache**: Redis for session storage and caching
- **Transactions**: Automatic transaction management with rollback support

### Security
- **Authentication**: Google OAuth2
- **Session Management**: Secure HTTP-only cookies
- **CSRF Protection**: Built into session management
- **Role-based Access**: Fine-grained permissions per endpoint

## Directory Structure

```
meals/
├── main.go                 # Application entry point
├── config/                 # Configuration management
│   ├── config.go          # Config struct and initialization
│   └── config.yaml        # Default configuration
├── routes/                 # Route definitions
│   └── routes.go          # All HTTP routes and middleware setup
├── handlers/               # HTTP request handlers
│   ├── auth.go            # Authentication handlers
│   ├── meal.go            # Meal CRUD operations
│   ├── menu.go            # Menu management
│   ├── profile.go         # User profile management
│   ├── home.go            # Home page handler
│   └── errors.go          # Standardized error handling
├── models/                 # Database models and business logic
│   ├── user.go            # User model with OAuth2 integration
│   ├── meal.go            # Meal model
│   ├── menu.go            # Menu model
│   └── user_profile.go    # User profile model
├── auth/                   # Authentication and authorization
│   ├── auth.go            # OAuth2 setup and session management
│   ├── role_auth.go       # Role-based authorization middleware
│   └── session.go         # Session middleware and management
├── middleware/             # HTTP middleware
│   ├── logger.go          # Request logging with request IDs
│   ├── recovery.go        # Panic recovery with logging
│   └── requestid.go       # Request ID generation and tracking
├── store/                  # Database connection and transaction management
│   ├── init.go            # Database initialization
│   ├── database.go        # PostgreSQL connection
│   ├── redis.go           # Redis connection
│   └── transaction.go     # Transaction management utilities
└── tests/                  # Test suites
    ├── models/            # Model tests
    └── testutils/         # Test utilities
```

## Data Flow

### Request Lifecycle
```
[HTTP Request] 
    ↓
[RequestID Middleware] → Generates unique request ID
    ↓
[Logger Middleware] → Logs request with ID
    ↓
[Recovery Middleware] → Handles panics gracefully
    ↓
[Authentication Check] → Validates session/OAuth2
    ↓
[Authorization Check] → Validates user roles
    ↓
[Route Handler] → Business logic execution
    ↓
[Database Transaction] → Data persistence with rollback
    ↓
[Response] → JSON response with request ID
```

### Authentication Flow
```
[User] → [/auth/google] → [Google OAuth2] → [Callback] → [Session Creation] → [Database User Record] → [Redirect to App]
```

### Database Transaction Flow
```
[Handler] → [store.WithTransaction()] → [Begin TX] → [Business Logic] → [Commit/Rollback] → [Response]
```

## Key Design Patterns

### Error Handling
- Standardized `ErrorResponse` struct
- Consistent error codes and messages
- Request ID tracking for debugging
- Automatic transaction rollback on errors

### Authentication
- Session-based authentication with OAuth2
- Role-based authorization middleware
- Automatic user creation on first login
- Secure session management

### Database Operations
- Transaction middleware for data integrity
- GORM ORM with automatic migrations
- Connection pooling and retry logic
- Separation of concerns (models vs handlers)

### Configuration
- Environment-based configuration
- YAML files with environment overrides
- Secure credential management
- Docker-friendly environment variables

## Security Considerations

### Authentication Security
- OAuth2 with Google (no password storage)
- Secure HTTP-only session cookies
- Session expiration and cleanup
- CSRF protection via session validation

### Database Security
- Parameterized queries (GORM prevents SQL injection)
- Transaction isolation
- Connection pooling with limits
- Credential management via environment variables

### API Security
- Role-based endpoint protection
- Request ID tracking for audit trails
- Panic recovery to prevent information leakage
- Structured error responses (no stack traces in production)

## Performance Considerations

### Database
- Connection pooling
- Redis caching for sessions
- Transaction management to prevent long-running locks
- Indexed queries for user lookups

### HTTP Server
- Gin framework for high performance
- Middleware pipeline optimization
- Request ID generation for tracing
- Graceful error handling

## Monitoring & Debugging

### Logging
- Structured logging with request IDs
- Request/response logging
- Error tracking with context
- Performance metrics (request duration)

### Error Tracking
- Panic recovery with stack traces
- Database error categorization
- Authentication failure logging
- Request ID correlation across logs 
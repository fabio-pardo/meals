# Project Navigation Map

## Quick Reference

### 🚀 Application Entry Points
- **`main.go:10`** - Application entry point and initialization sequence
- **`routes/routes.go:141`** - Router initialization and server startup
- **`routes/routes.go:14`** - All route definitions and middleware setup

### 🔐 Authentication Flow
1. **`routes/routes.go:67`** - Auth route definitions (`/auth/google`, `/logout`)
2. **`handlers/auth.go:15`** - OAuth2 handlers (login, callback, logout)
3. **`auth/auth.go:15`** - OAuth2 setup and session management
4. **`auth/role_auth.go:34`** - Role-based authorization middleware
5. **`auth/session.go:12`** - Session middleware and management

### 📊 Request Lifecycle
1. **`main.go:10`** - Application entry
2. **`routes/routes.go:141`** - Router initialization
3. **`middleware/requestid.go:16`** - Request ID generation
4. **`middleware/logger.go:10`** - Request logging
5. **`middleware/recovery.go:11`** - Panic recovery
6. **`auth/role_auth.go:34`** - Authentication/authorization
7. **`handlers/`** - Business logic execution
8. **`store/transaction.go:25`** - Database transaction management

### 🗄️ Database Operations
- **`store/init.go:2`** - Database initialization
- **`store/database.go:13`** - PostgreSQL connection setup
- **`store/redis.go:11`** - Redis connection setup
- **`store/transaction.go:25`** - Transaction management utilities
- **`models/`** - Database models and business logic

### 🍽️ Core Business Logic

#### Meals Management
- **`handlers/meal.go:12`** - Meal CRUD operations
- **`models/meal.go:7`** - Meal model definition
- **Routes**: `GET/POST /meals`, `GET/PUT/DELETE /meals/:id`

#### Menu Management
- **`handlers/menu.go:11`** - Menu CRUD operations
- **`models/menu.go:8`** - Menu model definition
- **`models/menu_meal.go:6`** - Menu-meal relationship model
- **Routes**: `GET/POST/PUT /menus`

#### User Profiles
- **`handlers/profile.go:25`** - User profile management
- **`models/user_profile.go:7`** - User profile model
- **Routes**: `GET/PUT /profile`, `PUT /profile/driver`

### ⚙️ Configuration Management
- **`config/config.go:61`** - Configuration initialization and structure
- **`config/config.yaml`** - Default configuration values
- **Environment Variables** - Override configuration in production

### 🛡️ Error Handling
- **`handlers/errors.go:38`** - Standardized error response system
- **`handlers/errors.go:197`** - Application error handling utilities
- **Error Codes**: `BAD_REQUEST`, `NOT_FOUND`, `UNAUTHORIZED`, etc.

### 🧪 Testing
- **`tests/models/`** - Model unit tests
- **`tests/testutils/db.go:10`** - Test database utilities
- **`tests/simple_test.go:13`** - Basic connectivity tests

## 🔍 Code Search Patterns

### Find All Functions
```bash
rg "^func [A-Z]" --type go
```

### Find All Routes
```bash
rg "router\.(GET|POST|PUT|DELETE)" --type go
```

### Find All Models
```bash
rg "^type.*struct" --type go
```

### Find Error Handling
```bash
rg "RespondWithError|HandleAppError" --type go
```

### Find Database Operations
```bash
rg "store\.DB\.|WithTransaction" --type go
```

## 📁 Directory Structure Guide

```
meals/
├── 📄 main.go                    # App entry point
├── 🔧 config/                    # Configuration management
│   ├── config.go                # Config struct and loading
│   └── config.yaml              # Default configuration
├── 🛣️ routes/                     # HTTP routing
│   └── routes.go                # All route definitions
├── 🎯 handlers/                   # HTTP request handlers
│   ├── auth.go                  # Authentication handlers
│   ├── meal.go                  # Meal CRUD operations
│   ├── menu.go                  # Menu management
│   ├── profile.go               # User profile management
│   ├── home.go                  # Home page handler
│   └── errors.go                # Error handling utilities
├── 📊 models/                     # Database models
│   ├── user.go                  # User model with OAuth2
│   ├── meal.go                  # Meal model
│   ├── menu.go                  # Menu model
│   ├── menu_meal.go             # Menu-meal junction
│   ├── user_profile.go          # User profile model
│   ├── session.go               # Session model
│   └── database.go              # Database wrapper
├── 🔐 auth/                       # Authentication & authorization
│   ├── auth.go                  # OAuth2 setup
│   ├── role_auth.go             # Role-based middleware
│   └── session.go               # Session management
├── 🔧 middleware/                 # HTTP middleware
│   ├── logger.go                # Request logging
│   ├── recovery.go              # Panic recovery
│   └── requestid.go             # Request ID tracking
├── 🗄️ store/                      # Database layer
│   ├── init.go                  # DB initialization
│   ├── database.go              # PostgreSQL setup
│   ├── redis.go                 # Redis setup
│   └── transaction.go           # Transaction utilities
├── 🧪 tests/                      # Test suites
│   ├── models/                  # Model tests
│   └── testutils/               # Test utilities
└── 📚 docs/                       # Documentation
    ├── architecture/            # System architecture
    ├── api/                     # API documentation
    ├── database/                # Database schema
    └── project-map.md           # This file
```

## 🔗 Key Relationships

### Authentication Chain
```
User Request → RequestID → Logger → Recovery → Auth Check → Role Check → Handler
```

### Database Transaction Flow
```
Handler → WithTransaction() → Begin TX → Business Logic → Commit/Rollback
```

### Error Handling Flow
```
Error Occurs → HandleAppError() → RespondWithError() → JSON Response with Request ID
```

### OAuth2 Flow
```
/auth/google → Google OAuth2 → /auth/google/callback → Session Creation → Redirect
```

## 🎯 Common Tasks

### Adding a New Endpoint
1. Define route in `routes/routes.go`
2. Create handler in appropriate `handlers/*.go` file
3. Add authentication/authorization if needed
4. Update API documentation in `docs/api/openapi.yaml`

### Adding a New Model
1. Create model struct in `models/*.go`
2. Add to database initialization in `store/database.go`
3. Create tests in `tests/models/*_test.go`
4. Update schema documentation in `docs/database/schema.md`

### Adding Middleware
1. Create middleware function in `middleware/*.go`
2. Add to middleware chain in `routes/routes.go:141`
3. Update request lifecycle documentation

### Debugging Issues
1. Check request ID in logs (generated by `middleware/requestid.go`)
2. Review error handling in `handlers/errors.go`
3. Check database transactions in `store/transaction.go`
4. Verify authentication flow in `auth/` directory

## 🔧 Development Workflow

### Local Development
1. **Setup**: `docker-compose up -d` (starts PostgreSQL + Redis)
2. **Run**: `go run main.go`
3. **Test**: `go test ./...`
4. **Docs**: View OpenAPI spec at `/docs` (if Swagger UI is added)

### Code Organization Principles
- **Separation of Concerns**: Clear boundaries between layers
- **Dependency Injection**: Database connections passed through context
- **Error Handling**: Consistent error responses with request IDs
- **Transaction Management**: Automatic rollback on errors
- **Authentication**: Session-based with role authorization

### Performance Monitoring
- **Request IDs**: Track requests across logs
- **Database Queries**: Monitor in `store/` layer
- **Error Rates**: Track in `handlers/errors.go`
- **Session Management**: Monitor in `auth/session.go` 
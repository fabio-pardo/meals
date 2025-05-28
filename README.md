# Meals App

A meal preparation service backend application built with Go. This application allows users to select upcoming meals for the week, make purchases, and have them delivered.

## Technologies Used

- Go (Golang)
- PostgreSQL for primary data storage
- Redis for caching
- OAuth2 (Google) for authentication
- Gin framework for API routing
- GORM for database ORM

## Configuration System

The app uses a structured configuration system with support for different environments:

- **development** - For local development
- **test** - For running tests
- **production** - For production deployment

### Configuration Files

Configuration is loaded in the following order (later sources override earlier ones):

1. Default values from the code
2. Base configuration from `config/config.yaml`
3. Environment-specific configuration from `config/config.[environment].yaml`
4. Environment variables (override everything else)

### Environment Variables

Key environment variables:

- `APP_ENV`: Sets the application environment (development, test, production)
- `DATABASE_*`: Database configuration
- `REDIS_*`: Redis configuration
- `AUTH_*`: Authentication configuration
- `SERVER_*`: Server configuration

## Getting Started

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd meals
   ```

2. **Set up development environment**
   ```bash
   make dev-setup
   ```

3. **Copy VS Code settings (optional)**
   ```bash
   cp .vscode/settings.json.template .vscode/settings.json
   ```

4. **Start the application**
   ```bash
   make dev
   ```

This will start PostgreSQL, Redis, and the application automatically.

### Manual Setup

1. **Install dependencies**
   ```bash
   go mod tidy
   ```

2. **Start databases**
   ```bash
   make docker-up
   ```

3. **Run the application**
   ```bash
   make run
   ```

### Development Commands

Run `make help` to see all available commands:

```bash
make help              # Show all commands
make dev-setup         # Set up development environment
make run               # Run the application
make test              # Run tests
make search-funcs      # Find all functions
make search-routes     # Find all routes
make docs              # Generate API documentation
```

## API Endpoints

### Authentication

- `GET /auth/google`: Start Google OAuth2 authentication
- `GET /auth/google/callback`: Google OAuth2 callback URL
- `GET /logout`: Log out the current user

### Meals

- `GET /meals`: List all meals
- `POST /meals`: Create a new meal
- `GET /meals/:id`: Get a specific meal
- `PUT /meals/:id`: Update a meal
- `DELETE /meals/:id`: Delete a meal

### Menus

- `POST /menus`: Create a new menu
- `PUT /menus`: Update a menu

## Docker Deployment

The application includes Docker and Docker Compose configurations for easy deployment.

```bash
# Build and start all containers
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all containers
docker-compose down
```

## Environment Variables for Docker

You can create a `.env` file in the project root to configure the Docker deployment:

```
APP_ENV=production
DATABASE_USER=meals_app_user
DATABASE_PASSWORD=strong_password
DATABASE_NAME=meals_production_db
AUTH_GOOGLEKEY=your_google_client_id
AUTH_GOOGLESECRET=your_google_client_secret
AUTH_GOOGLEREDIRECTURL=https://your-domain.com/auth/google/callback
AUTH_SESSIONSECRET=your_session_secret_key
```

## Documentation

- **Architecture**: `docs/architecture/README.md` - System design and components
- **API Reference**: `docs/api/openapi.yaml` - OpenAPI 3.1 specification
- **Database Schema**: `docs/database/schema.md` - Database structure and relationships
- **Project Navigation**: `docs/project-map.md` - Quick reference for code locations

## Development Workflow

### Code Search & Navigation

```bash
# Find all functions
make search-funcs

# Find all routes
make search-routes

# Find all models
make search-models

# Find error handling patterns
make search-errors

# Find database operations
make search-db
```

### Testing

```bash
make test              # Run all tests
make test-verbose      # Run with verbose output
make test-coverage     # Generate coverage report
```

### Code Quality

```bash
make lint              # Run linter
make format            # Format code
make vet               # Run go vet
```

## Project Structure

```
meals/
├── docs/              # Documentation
├── handlers/          # HTTP request handlers
├── models/            # Database models
├── auth/              # Authentication & authorization
├── middleware/        # HTTP middleware
├── store/             # Database layer
├── config/            # Configuration management
├── routes/            # Route definitions
└── tests/             # Test suites
```

## License

This project is licensed under the MIT License. 
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

### Local Development

1. Clone the repository
2. Copy the configuration template:
   ```
   cp config/config.yaml config/config.development.yaml
   ```
3. Edit `config/config.development.yaml` with your settings
4. Run with Docker Compose:
   ```
   docker-compose up -d
   ```

### Manual Setup

1. Install PostgreSQL and Redis
2. Set up the configuration as described above
3. Run the application:
   ```
   go run main.go
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

## License

This project is licensed under the MIT License. 
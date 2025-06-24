# P4rsec Go Server

A high-performance Go server built with Fiber framework, featuring PostgreSQL database, Redis caching, and comprehensive configuration management.

## Features

- 🚀 **High Performance**: Built with Fiber v2 framework
- 🗄️ **Database**: PostgreSQL with connection pooling
- 🔄 **Caching**: Redis for high-performance caching
- ⚙️ **Configuration**: Viper-based configuration with environment support
- 🏗️ **Architecture**: Clean architecture with DAO pattern
- 🔐 **Security**: Helmet middleware for security headers
- 📊 **Logging**: Structured logging with Zap
- 🐳 **Docker**: Production-ready containerization
- 🧪 **Testing**: Comprehensive test setup
- 📈 **Health Checks**: Built-in health monitoring

## Project Structure

```
server/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   ├── postgres.go          # PostgreSQL connection
│   │   └── redis.go             # Redis connection
│   ├── dao/
│   │   ├── user_dao.go          # User data access layer
│   │   └── cache_dao.go         # Cache operations
│   ├── handlers/
│   │   ├── health_handler.go    # Health check endpoints
│   │   └── user_handler.go      # User CRUD operations
│   ├── logger/
│   │   └── logger.go            # Structured logging
│   ├── models/
│   │   └── user.go              # Data models
│   └── server/
│       └── server.go            # Server setup and middleware
├── configs/
│   ├── config.yaml              # Base configuration
│   ├── config.development.yaml  # Development config
│   ├── config.staging.yaml      # Staging config
│   └── config.production.yaml   # Production config
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   └── 000001_create_users_table.down.sql
├── docker-compose.yml           # Development environment
├── Dockerfile                   # Production container
├── Makefile                     # Build and development tasks
├── go.mod                       # Go module definition
└── .env                         # Environment variables
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)

### Development Setup

1. **Clone and setup the project:**

   ```bash
   cd server
   make deps
   ```

2. **Start development services:**

   ```bash
   make dev-env
   ```

3. **Run database migrations:**

   ```bash
   make migrate-up
   ```

4. **Start the server:**
   ```bash
   make run
   ```

The server will be available at `http://localhost:8080`

### Using Docker

1. **Start all services:**

   ```bash
   docker-compose up -d
   ```

2. **Build and run the application:**
   ```bash
   make docker-build
   make docker-run
   ```

## Configuration

The application supports multiple environments with hierarchical configuration:

### Environment Variables

All configuration can be overridden with environment variables using the `APP_` prefix:

```bash
export APP_SERVER_PORT=8080
export APP_DATABASE_HOST=localhost
export APP_REDIS_HOST=localhost
```

### Configuration Files

- `config.yaml` - Base configuration
- `config.development.yaml` - Development overrides
- `config.staging.yaml` - Staging overrides
- `config.production.yaml` - Production overrides

### Environment Setup

Set the environment using:

```bash
export APP_ENVIRONMENT=development  # or staging, production
```

## API Endpoints

### Health Check

- `GET /api/v1/health` - Service health status

### Users

- `GET /api/v1/users` - List users (with pagination)
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user (soft delete)

### Example API Usage

```bash
# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe"
  }'

# Get users with pagination
curl "http://localhost:8080/api/v1/users?page=1&limit=10"

# Get user by ID
curl http://localhost:8080/api/v1/users/uuid-here

# Update user
curl -X PUT http://localhost:8080/api/v1/users/uuid-here \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane"
  }'
```

## Database

### Migrations

The project uses `golang-migrate` for database migrations:

```bash
# Create new migration
make migrate-create name=add_user_table

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

### Schema

The current schema includes:

- **users** table with UUID primary keys
- Optimized indexes for common queries
- Soft delete functionality

## Caching Strategy

The application implements a multi-layer caching strategy:

1. **Individual Records**: Cache frequently accessed user records
2. **List Queries**: Cache paginated user lists
3. **Cache Invalidation**: Automatic cache invalidation on updates
4. **Session Management**: Redis-based session storage

## Development Commands

```bash
# Build the application
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Security scan
make security

# Start development environment
make dev-env

# Stop development environment
make dev-env-stop

# Install development tools
make install-tools
```

## Production Deployment

### Docker Deployment

```bash
# Build production image
make docker-build

# Run in production mode
docker run -p 8080:8080 \
  --env-file .env.production \
  p4rsec-server
```

### Environment Variables for Production

Set the following environment variables in production:

```bash
APP_ENVIRONMENT=production
APP_SERVER_PORT=8080
APP_DATABASE_HOST=your-db-host
APP_DATABASE_PASSWORD=your-secure-password
APP_REDIS_HOST=your-redis-host
APP_REDIS_PASSWORD=your-redis-password
APP_JWT_SECRET=your-jwt-secret
```

## Security Features

- **CORS**: Configurable cross-origin resource sharing
- **Helmet**: Security headers middleware
- **Rate Limiting**: Request rate limiting per IP
- **Input Validation**: Request body validation
- **SQL Injection Prevention**: Parameterized queries
- **Non-root Container**: Runs as non-privileged user

## Monitoring and Observability

- **Health Checks**: Built-in health endpoints
- **Structured Logging**: JSON-formatted logs in production
- **Connection Pooling**: Database connection pool monitoring
- **Graceful Shutdown**: Proper resource cleanup on shutdown

## Performance Optimizations

- **Connection Pooling**: Optimized database connections
- **Redis Caching**: High-performance caching layer
- **Batch Operations**: Efficient bulk operations
- **Query Optimization**: Indexed database queries

## Contributing

1. Follow Go best practices and conventions
2. Add tests for new functionality
3. Update documentation for API changes
4. Use conventional commit messages
5. Ensure all tests pass before submitting

## License

This project is licensed under the MIT License.

# Auth Microservice Go

A robust, production-ready authentication microservice built with Go, featuring JWT tokens, role-based access control (RBAC), and comprehensive observability.

## 🚀 Features

- **JWT Authentication**: Secure token-based authentication with access/refresh token pairs
- **Role-Based Access Control (RBAC)**: Flexible permission system with roles and scopes
- **Client Credentials Flow**: OAuth2-style client authentication for service-to-service communication
- **Token Management**: Token introspection, rotation, and revocation
- **Redis Caching**: High-performance caching for tokens and user sessions
- **PostgreSQL Database**: Reliable data persistence with GORM ORM
- **Swagger Documentation**: Auto-generated API documentation
- **Observability**: Comprehensive monitoring with Prometheus and Grafana
- **Health Checks**: Built-in health monitoring endpoints
- **Docker Support**: Containerized deployment with Docker Compose
- **Centralized Configuration**: Environment-based configuration with validation
- **Structured Error Handling**: Consistent JSON error responses with proper HTTP status codes
- **Production Ready**: Clean architecture, logging, metrics, and tracing

## 🏗️ Architecture

The project follows Clean Architecture principles with clear separation of concerns:

```
├── cmd/auth-service/          # Application entry point
├── internal/
│   ├── domain/               # Business entities and interfaces
│   ├── usecase/              # Business logic implementation
│   ├── config/               # Centralized configuration management
│   ├── errors/               # Structured error handling system
│   ├── infra/                # Infrastructure layer
│   │   ├── db/               # Database repositories and models
│   │   ├── cache/            # Redis caching implementation
│   │   ├── logger/           # Structured logging configuration
│   │   ├── metrics/          # Prometheus metrics
│   │   └── trace/            # OpenTelemetry tracing
│   └── transport/            # HTTP handlers and middleware
├── docs/                     # Swagger documentation
└── observability/            # Monitoring configuration (Grafana, Prometheus)
```

## 🛠️ Tech Stack

- **Language**: Go 1.24+
- **Web Framework**: Chi Router
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **ORM**: GORM
- **Authentication**: JWT with golang-jwt/jwt/v5
- **Documentation**: Swagger/OpenAPI
- **Monitoring**: Prometheus + Grafana
- **Validation**: go-playground/validator
- **Containerization**: Docker & Docker Compose

## 📋 Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local development)
- PostgreSQL 15+ (if running locally)
- Redis 7+ (if running locally)

## 🚦 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone https://github.com/YuriGarciaRibeiro/auth-microservice-go.git
   cd auth-microservice-go
   ```

2. **Set up environment variables**
   ```bash
   # Copy the example environment file
   cp .env.example .env
   
   # Edit the .env file with your configurations
   ```

3. **Start all services**
   ```bash
   docker-compose up -d --build
   ```

4. **Verify the setup**
   ```bash
   # Check if all services are running
   docker-compose ps
   
   # Test the health endpoint
   curl http://localhost:8080/healthz
   ```

### Local Development

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Set up the database**
   ```bash
   # Start PostgreSQL and Redis
   docker-compose up -d postgres redis
   
   # Run database migrations (if available)
   # migrate -path migrations -database "postgres://user:pass@localhost:5432/auth_db?sslmode=disable" up
   ```

3. **Set environment variables**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=user
   export DB_PASSWORD=pass
   export DB_NAME=auth_db
   export REDIS_ADDR=localhost:6379
   export ACCESS_SECRET=your-access-secret
   export REFRESH_SECRET=your-refresh-secret
   ```

4. **Run the application**
   ```bash
   go run cmd/auth-service/main.go
   ```

## 🔧 Configuration

The application now features **centralized configuration** with automatic validation and clear error messages.

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_PORT` | Server port | `8080` | ❌ |
| `SERVER_HOST` | Server host | `` | ❌ |
| `DB_HOST` | PostgreSQL host | `localhost` | ✅ |
| `DB_PORT` | PostgreSQL port | `5432` | ✅ |
| `DB_USER` | PostgreSQL user | `user` | ✅ |
| `DB_PASSWORD` | PostgreSQL password | - | ✅ |
| `DB_NAME` | PostgreSQL database name | `auth_db` | ✅ |
| `REDIS_ADDR` | Redis address | `localhost:6379` | ❌ |
| `REDIS_PASS` | Redis password | - | ❌ |
| `REDIS_DB` | Redis database number | `0` | ❌ |
| `ACCESS_SECRET` | JWT access token secret | - | ✅ |
| `REFRESH_SECRET` | JWT refresh token secret | - | ✅ |
| `ACCESS_TOKEN_TTL` | Access token TTL | `15m` | ❌ |
| `REFRESH_TOKEN_TTL` | Refresh token TTL | `168h` (7 days) | ❌ |
| `JWT_ISSUER` | JWT token issuer | `auth-microservice` | ❌ |
| `JWT_AUDIENCE` | JWT token audience (CSV) | - | ❌ |
| `CACHE_PROFILE_TTL` | Profile cache TTL | `5m` | ❌ |
| `PERM_CACHE_TTL` | Permission cache TTL | `15m` | ❌ |
| `LOG_LEVEL` | Log level (debug,info,warn,error) | `info` | ❌ |
| `LOG_ENCODING` | Log encoding (json,console) | `json` | ❌ |
| `APP_ENV` | Application environment | `dev` | ❌ |

### Configuration Validation

The application validates all required configuration on startup and provides clear error messages:

```bash
# Missing required configuration
Configuration error: ACCESS_SECRET is required

# Invalid configuration format  
Configuration error: invalid duration format for ACCESS_TOKEN_TTL
```

## 📚 API Documentation

The API documentation is automatically generated using Swagger and available at:
- **Swagger UI**: http://localhost:8080/docs/
- **OpenAPI JSON**: http://localhost:8080/docs/swagger.json
- **OpenAPI YAML**: http://localhost:8080/docs/swagger.yaml

### 🎯 Structured Error Responses

All API endpoints now return consistent, structured error responses in JSON format:

```json
{
  "type": "validation_error",
  "message": "Validation failed", 
  "details": "Field validation for 'Email' failed on the 'email' tag"
}
```

**Error Types:**
- `validation_error` (422) - Request validation failed
- `authentication_error` (401) - Invalid credentials or token
- `authorization_error` (403) - Insufficient permissions
- `conflict_error` (409) - Resource already exists
- `not_found` (404) - Resource not found
- `bad_request` (400) - Malformed request
- `internal_error` (500) - Server error

### 🔐 Authentication Endpoints

#### User Authentication
- `POST /auth/signup` - Register a new user
- `POST /auth/login` - Authenticate user and get tokens
- `POST /auth/logout` - Revoke tokens and logout
- `POST /auth/refresh` - Refresh access token using refresh token
- `POST /auth/introspect` - Validate and introspect access token

#### Client Authentication (OAuth2 Client Credentials)
- `POST /auth/token` - Get access token using client credentials

### 👨‍💼 Admin Endpoints (Protected)

#### Scope Management
- `POST /admin/scopes` - Create new scope
- `GET /admin/scopes` - List all scopes

#### Role Management
- `POST /admin/roles` - Create new role
- `GET /admin/roles` - List all roles
- `POST /admin/roles/{roleId}/scopes` - Attach scopes to role

#### User Management
- `POST /admin/users/{userId}/roles` - Assign roles to user
- `GET /admin/users/{userId}/roles` - Get user roles
- `GET /admin/users/{userId}/scopes` - Get user effective scopes
- `POST /admin/users/{userId}/scopes/grant` - Grant direct scope to user
- `POST /admin/users/{userId}/scopes/revoke` - Revoke direct scope from user

#### Client Management
- `POST /admin/clients/{clientId}/scopes` - Assign scopes to client
- `GET /admin/clients/{clientId}/scopes` - Get client scopes

### 📊 Monitoring Endpoints
- `GET /healthz` - Health check endpoint
- `GET /metrics` - Prometheus metrics

## 🔒 Security Features

### JWT Token Security
- **Separate secrets** for access and refresh tokens
- **Short-lived access tokens** (15 minutes default)
- **Long-lived refresh tokens** (7 days default) stored in Redis
- **Token blacklisting** for logout functionality
- **Token rotation** on refresh

### Role-Based Access Control (RBAC)
- **Flexible permission system** with roles and scopes
- **Fine-grained access control** at endpoint level
- **Dynamic permission assignment** through admin endpoints
- **Effective permissions** calculation (roles + direct scopes)

### Password Security
- **BCrypt hashing** with appropriate cost factor
- **Password validation** with minimum length requirements

## 📈 Monitoring & Observability

### Prometheus Metrics
The service exposes various metrics for monitoring:
- HTTP request duration and count
- Database connection metrics
- Redis operation metrics
- Custom business metrics

Access Prometheus at: http://localhost:9090

### Grafana Dashboards
Pre-configured dashboards for:
- Service overview and health
- HTTP request metrics
- Database performance
- Cache performance

Access Grafana at: http://localhost:3000 (admin/admin)

### Logging
Structured logging with:
- Request/response logging
- Error tracking
- Performance metrics
- Security events

## 🗄️ Database Schema

### Core Tables
- **users**: User accounts and credentials
- **clients**: OAuth2 clients for service-to-service auth
- **roles**: Permission roles
- **scopes**: Permission scopes
- **user_roles**: User-role assignments
- **role_scopes**: Role-scope assignments
- **user_scopes**: Direct user-scope assignments
- **client_scopes**: Client-scope assignments

## 🧪 Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### API Testing Examples

#### User Registration
```bash
curl -X POST "http://localhost:8080/auth/signup" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "123456"
  }'
```

**Success Response (201):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "access_exp": "2025-08-19T15:30:00Z",
  "refresh_exp": "2025-08-26T14:30:00Z"
}
```

**Error Response - Validation Failed (422):**
```json
{
  "type": "validation_error",
  "message": "Validation failed",
  "details": "Field validation for 'Email' failed on the 'email' tag"
}
```

**Error Response - User Already Exists (409):**
```json
{
  "type": "conflict_error",
  "message": "User with this email already exists"
}
```

#### User Login
```bash
curl -X POST "http://localhost:8080/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "123456"
  }'
```

**Error Response - Invalid Credentials (401):**
```json
{
  "type": "authentication_error",
  "message": "Invalid email or password"
}
```

#### Client Credentials Flow
```bash
curl -X POST "http://localhost:8080/auth/token" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "service-a",
    "client_secret": "service-a-secret",
    "scopes": ["read:users"],
    "audience": ["service-b"]
  }'
```

#### Token Introspection
```bash
curl -X POST "http://localhost:8080/auth/introspect" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your-access-token-here"
  }'
```

#### Testing Protected Endpoints
```bash
# Missing Authorization header
curl -X GET "http://localhost:8080/admin/scopes"

# Response (401):
{
  "type": "authentication_error",
  "message": "Missing or malformed Authorization header"
}

# Invalid token
curl -X GET "http://localhost:8080/admin/scopes" \
  -H "Authorization: Bearer invalid-token"

# Response (401):
{
  "type": "authentication_error", 
  "message": "Invalid or expired token"
}

# Valid token but insufficient permissions
curl -X GET "http://localhost:8080/admin/scopes" \
  -H "Authorization: Bearer <valid-user-token-without-admin-role>"

# Response (403):
{
  "type": "authorization_error",
  "message": "Insufficient permissions: missing required role"
}
```

## 🐳 Docker Services

The docker-compose.yml includes:

- **auth-service**: The main authentication service
- **postgres**: PostgreSQL database  
- **redis**: Redis cache
- **pgadmin**: PostgreSQL administration interface (http://localhost:5050)
- **prometheus**: Metrics collection (http://localhost:9090)
- **grafana**: Metrics visualization (http://localhost:3000)
- **jaeger**: Distributed tracing interface (http://localhost:16686)

## 🚀 Deployment

### Production Considerations

1. **Environment Variables**: Set strong secrets for JWT tokens
2. **Database**: Use managed PostgreSQL service in production
3. **Cache**: Use managed Redis service in production
4. **Monitoring**: Set up alerts for critical metrics
5. **Logging**: Configure log aggregation (ELK stack, etc.)
6. **Backup**: Implement database backup strategy
7. **SSL/TLS**: Use reverse proxy with SSL termination

### Kubernetes Deployment
For Kubernetes deployment, consider:
- ConfigMaps for configuration
- Secrets for sensitive data
- Horizontal Pod Autoscaler for scaling
- Persistent Volumes for data storage
- Service mesh for secure communication

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Write tests for new features
- Update documentation for API changes
- Ensure all tests pass before submitting PR
- Use meaningful commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**Yuri Garcia Ribeiro**
- GitHub: [@YuriGarciaRibeiro](https://github.com/YuriGarciaRibeiro)
- Project: [auth-microservice-go](https://github.com/YuriGarciaRibeiro/auth-microservice-go)

## 🙏 Acknowledgments

- [Chi Router](https://github.com/go-chi/chi) for the HTTP routing
- [GORM](https://gorm.io/) for the ORM
- [JWT-Go](https://github.com/golang-jwt/jwt) for JWT implementation
- [Prometheus](https://prometheus.io/) for metrics
- [Grafana](https://grafana.com/) for visualization

## 📈 Project Status

**Production Ready** ✅

This microservice has been thoroughly reviewed and enhanced with:
- ✅ Centralized configuration with validation
- ✅ Structured error handling across all endpoints
- ✅ Clean architecture and code organization
- ✅ Comprehensive observability (logs, metrics, tracing)
- ✅ Docker containerization
- ✅ API documentation with Swagger
- ✅ JWT-based authentication with RBAC
- ✅ Redis caching for performance
- ✅ PostgreSQL for reliable data persistence

**Latest Version**: Enhanced with structured error responses and centralized configuration management.

⭐ **Star this repository if you find it helpful!**

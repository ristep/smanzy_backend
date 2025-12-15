# User Management API with JWT Authentication

A production-ready REST API in Go featuring secure user management, role-based access control (RBAC), and JWT authentication.

## Tech Stack

- **Web Framework**: [Gin](https://github.com/gin-gonic/gin) - High-performance HTTP web framework
- **JWT Library**: [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) - JWT authentication
- **ORM**: [GORM](https://gorm.io/) - Object-relational mapping with auto-migrations
- **Password Hashing**: [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - Secure password storage
- **Database**: PostgreSQL (configurable via connection string)
- **Environment**: [godotenv](https://github.com/joho/godotenv) - Environment variable management

## Project Structure

```
um_starter_jwt_go/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── models/
│   │   └── user.go                 # User and Role data models
│   ├── handlers/
│   │   └── auth.go                 # HTTP handlers for auth and user management
│   ├── middleware/
│   │   └── auth.go                 # JWT and RBAC middleware
│   └── auth/
│       └── jwt.go                  # JWT token generation and validation
├── go.mod                           # Go module dependencies
├── go.sum                           # Go module checksums
├── .env.example                     # Example environment variables
└── README.md                        # This file
```

## Setup Instructions

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Git

### 1. Clone or Initialize the Project

```bash
cd /path/to/um_starter_jwt_go
```

### 2. Download Dependencies

```bash
go mod download
```

### 3. Configure Environment Variables

Copy the example environment file and configure it:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Database Configuration
DB_DSN=postgres://user:password@localhost:5432/smanzy_db?sslmode=disable

# JWT Configuration (use a strong random key)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Server Configuration
SERVER_PORT=8080
```

**Generate a secure JWT secret:**

```bash
openssl rand -base64 32
```

### 4. Create PostgreSQL Database

```bash
createdb smanzy_db
```

Or use PostgreSQL client:

```sql
CREATE DATABASE smanzy_db;
```

### 5. Run the Application

```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

### Optional: pgAdmin

You can run pgAdmin as a Docker container (this repository's `docker-compose.yml` includes a `pgadmin` service):

1. Start the containers:

```bash
docker-compose up -d
```

2. Visit pgAdmin:

  - URL: `http://localhost:5050`
  - **Default pgAdmin credentials (from docker-compose)**:
    - Email: `ristep@gmail.com`
    - Password: `Leprakon!2025`

3. Add a Server to pgAdmin (recommended):

  - General:
    - Name: `smanzy_postgres` (or any friendly name)
  - Connection:
    - Host name/address: `postgres` (Required when running in Docker)
      (Do NOT use `localhost` inside Docker network)
    - Port: `5432`
    - Username: `postgres`
    - Password: `postgres`
    - Maintenance DB: `postgres` or `smanzy_db`
    - Save password: ✓

4. Pre-configured servers file

  - A sample `servers.json` is included at `docker/pgadmin/servers.json`. This is mounted into the container at `/pgadmin4/servers.json`, so pgAdmin will show the `smanzy_postgres` server by default.
  - If you change the postgres service name in `docker-compose.yml`, be sure to update `docker/pgadmin/servers.json` as well.

5. Restart to pick up changes and server config:

```bash
docker-compose down && docker-compose up -d
```

> Note: The `pgadmin` service mounts `pgadmin_data` for persistent pgAdmin storage, so your configured servers will persist between container restarts.

  > Security note: For security reasons, the preconfigured `servers.json` does not include the database password; you should enter the DB password in the pgAdmin UI. Avoid committing actual credentials to version control. Add any personal secret values to your local `.env` and keep it out of the repo.

## API Endpoints

### Health Check

```
GET /health
Response: {"status": "ok"}
```

### Public Endpoints

#### Register a New User

```
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "John Doe"
}

Response (201 Created):
{
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe",
      "roles": [{"id": 1, "name": "user"}],
      "created_at": 1702324800000
    },
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  }
}
```

#### Login

```
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}

Response (200 OK):
{
  "data": {
    "user": {...},
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  }
}
```

#### Refresh Token

```
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGc..."
}

Response (200 OK):
{
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  }
}

#### List Public Media

```
GET /api/media?limit=100&offset=0

Response (200 OK):
{
  "data": [
    {
      "id": 1,
      "filename": "image.jpg",
      "url": "/api/media/files/1_1600000000.jpg",
      "type": "file",
      "mime_type": "image/jpeg",
      "size": 12345,
      "created_at": 1702324800000
    }
  ]
}
```

Notes:
- `limit` and `offset` query parameters are supported for pagination. Defaults: `limit=100`, `offset=0`.
- This endpoint is public (no authentication required) and returns basic metadata only.
```

### Protected Endpoints

All protected endpoints require the `Authorization` header:

```
Authorization: Bearer <access_token>
```

#### Get User Profile

```
GET /api/profile
Authorization: Bearer <access_token>

Response (200 OK):
{
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "roles": [{"id": 1, "name": "user"}],
    "created_at": 1702324800000
  }
}
```

### Admin-Only Endpoints

Requires `Authorization: Bearer <access_token>` and `admin` role.

#### Get All Users

```
GET /api/users
Authorization: Bearer <admin_token>

Response (200 OK):
{
  "data": [
    {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe",
      "roles": [{"id": 1, "name": "user"}],
      "created_at": 1702324800000
    }
  ]
}
```

#### Get User by ID

```
GET /api/users/:id
Authorization: Bearer <admin_token>

Response (200 OK):
{
  "data": {...}
}
```

#### Update User

```
PUT /api/users/:id
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Updated Name"
}

Response (200 OK):
{
  "data": {...}
}
```

#### Delete User

```
DELETE /api/users/:id
Authorization: Bearer <admin_token>

Response (200 OK):
{
  "data": {"message": "User deleted successfully"}
}
```

#### Assign Role to User

```
POST /api/users/:id/roles
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "role_name": "admin"
}

Response (200 OK):
{
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "roles": [
      {"id": 1, "name": "user"},
      {"id": 2, "name": "admin"}
    ],
    "created_at": 1702324800000
  }
}
```

#### Remove Role from User

```
DELETE /api/users/:id/roles
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "role_name": "admin"
}

Response (200 OK):
{
  "data": {...}
}
```

## Authentication Flow

1. **Registration**: User registers with email, password, and name
   - Password is hashed using bcrypt
   - User is assigned the default "user" role
   - Access and refresh tokens are returned

2. **Login**: User logs in with email and password
   - Credentials are validated
   - Bcrypt comparison verifies password
   - Tokens are generated and returned

3. **Token Usage**: User includes access token in `Authorization: Bearer <token>` header
   - Access tokens are short-lived (15 minutes)
   - Middleware validates token signature and expiration

4. **Token Refresh**: User uses refresh token to get new access token
   - Refresh tokens are long-lived (7 days)
   - New access token is issued without re-authentication

## Role-Based Access Control (RBAC)

The API implements role-based access control with the following default roles:

- **user**: Standard user role (default for new registrations)
- **admin**: Administrator with full access to user management endpoints

### Role Middleware

The `RoleMiddleware` checks if a user has at least one of the allowed roles:

```go
users.Use(middleware.RoleMiddleware("admin"))
```

Multiple roles can be specified:

```go
users.Use(middleware.RoleMiddleware("admin", "moderator"))
```

## Security Features

### Password Security

- Passwords are hashed using bcrypt with default cost (10 rounds)
- Password comparisons use bcrypt's timing-safe comparison
- Passwords are never logged or exposed in API responses

### JWT Security

- Tokens are signed using HMAC-SHA256
- JWT secret is loaded from environment variables (never hardcoded)
- Token validation checks:
  - Signature verification
  - Expiration time
  - Signing method (prevents algorithm confusion attacks)

### Database Security

- User model uses GORM soft deletes for audit trail
- Password field is excluded from JSON serialization (`json:"-"`)
- Email field is unique at database level

### Middleware Security

- `AuthMiddleware` extracts tokens from `Authorization: Bearer <token>` header
- Invalid tokens return 401 Unauthorized
- Missing authorization headers return 401 Unauthorized
- Insufficient permissions return 403 Forbidden

## Development

### Code Organization

- **cmd/**: Executable code (main.go and entry points)
- **internal/**: Private package code
  - **models/**: GORM data models
  - **handlers/**: HTTP request handlers
  - **middleware/**: Gin middleware functions
  - **auth/**: JWT token logic

### Naming Conventions

- Handlers: `*Handler` suffix
- Services: `*Service` suffix
- Middleware: Returns `gin.HandlerFunc`
- Requests: `*Request` suffix
- Responses: `*Response` suffix or use `SuccessResponse`/`ErrorResponse`

### Error Handling

- Validation errors: 400 Bad Request
- Authentication errors: 401 Unauthorized
- Authorization errors: 403 Forbidden
- Not found: 404 Not Found
- Server errors: 500 Internal Server Error

## Production Deployment

### Environment Variables

Ensure these are set in your production environment:

```bash
export DB_DSN="postgres://user:password@prod-db:5432/smanzy_db?sslmode=require"
export JWT_SECRET="$(openssl rand -base64 32)"
export SERVER_PORT="8080"
export ENV="production"
```

### Database Migrations

Migrations are automatically run on startup via `AutoMigrate()`. No manual migration steps required.

### TLS/HTTPS

For production, deploy behind a reverse proxy (nginx, Caddy) that handles TLS.

### Logging

Currently uses `log` package. For production, consider:

- [Zap](https://github.com/uber-go/zap) - Structured logging
- [Logrus](https://github.com/sirupsen/logrus) - Structured logging with levels

### Performance Tuning

- Database connection pooling is configured in GORM
- Gin runs in release mode in production (set `gin.SetMode(gin.ReleaseMode)`)
- Use reverse proxy caching for `/health` endpoint

## Testing

(To be implemented)

### Unit Tests

```bash
go test ./...
```

### Integration Tests

Set `TEST_DB_DSN` environment variable and run:

```bash
go test ./... -v
```

## Contributing

1. Follow Go code style guidelines
2. Add tests for new features
3. Document API changes
4. Never commit sensitive data (.env files, tokens, etc.)

## License

MIT License - See LICENSE file for details

## Support

For issues, questions, or contributions, please open an issue or pull request.

# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

A production-ready REST API in Go featuring JWT authentication, role-based access control (RBAC), user management, and media file handling. Built with Gin web framework, GORM ORM, and PostgreSQL.

## Development Commands

### Running the Application
```bash
# Run directly
go run cmd/api/main.go

# Using Make
make run

# Development with hot-reload (requires air)
make dev
```

### Building
```bash
make build  # Creates binary in ./bin/smanzy
```

### Testing
```bash
# Run all tests
make test

# Test with coverage
go test -v -cover ./...
```

### Code Quality
```bash
make fmt   # Format code with go fmt
make lint  # Lint code (requires golangci-lint)
```

### Dependencies
```bash
make deps  # Download and tidy dependencies
```

### Database
```bash
# Start PostgreSQL and pgAdmin via Docker
docker-compose up -d

# pgAdmin is accessible at http://localhost:5050
```

## Architecture

### Project Structure
```
cmd/api/main.go          - Application entry point, routing, DB initialization
internal/
  models/                - Data models (User, Role, Media)
  handlers/              - HTTP request handlers (auth, user, media)
  middleware/            - Middleware (JWT auth, RBAC, CORS)
  auth/                  - JWT service (token generation/validation)
uploads/                 - Media file storage directory
```

### Key Architectural Patterns

**Dependency Injection**: Services (DB, JWTService) are initialized in `main.go` and injected into handlers
```go
jwtService := auth.NewJWTService(jwtSecret)
authHandler := handlers.NewAuthHandler(db, jwtService)
```

**Middleware Chain**: Authentication and authorization are handled via middleware
- `AuthMiddleware`: Validates JWT and attaches user to context
- `RoleMiddleware`: Checks user roles for authorization
- Apply to route groups in `main.go`

**Context-Based User Access**: Authenticated user is stored in Gin context
```go
user, _ := c.Get("user")
userObj := user.(*models.User)
```

**GORM Patterns**:
- Auto-migration on startup via `db.AutoMigrate()`
- Soft deletes using `gorm.DeletedAt`
- Many-to-many relationships (User-Role) via join table
- Preloading associations: `db.Preload("Roles").First(&user, id)`

### JWT Token System
- **Access tokens**: 15-minute lifespan, used for API requests
- **Refresh tokens**: 7-day lifespan, used to obtain new access tokens
- Both tokens contain: UserID, Email, Name, Roles
- Issuer: `"um-api"`
- Signing method: HS256

### Role-Based Access Control
- Default roles: `user`, `admin`
- Roles are seeded on application startup
- New users automatically get `user` role
- Admin routes protected by `RoleMiddleware("admin")`
- Check roles programmatically: `user.HasRole("admin")`

### Media File Handling
- Files uploaded via multipart/form-data to `/api/media`
- Stored in `./uploads/` directory with unique names
- Metadata tracked in database (Media model)
- Access control: users can delete/edit their own files, admins can manage all

## Environment Variables

Required variables in `.env`:
```bash
DB_DSN=postgres://user:pass@localhost:5432/dbname?sslmode=disable
JWT_SECRET=<generated-secret>  # Generate: openssl rand -base64 32
SERVER_PORT=8080
ENV=development
```

## Database Schema

**users**: id, email (unique), password (hashed), name, tel, age, address, city, country, gender, email_verified, created_at, updated_at, deleted_at

**roles**: id, name (unique), created_at, updated_at

**user_roles**: user_id, role_id (join table)

**media**: id, filename, stored_name, url, type, mime_type, size, user_id, created_at, updated_at, deleted_at

## API Routes

### Public
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login and get tokens
- `POST /api/auth/refresh` - Refresh access token
- `GET /api/media` - List public media
- `GET /api/media/files/:name` - Serve files (development)
- `GET /health` - Health check

### Protected (JWT required)
- `GET /api/profile` - Current user profile
- `PUT /api/profile` - Update profile
- `POST /api/media` - Upload file
- `GET /api/media/:id` - Get media
- `GET /api/media/:id/details` - Get media metadata
- `PUT /api/media/:id` - Update media (owner/admin)
- `DELETE /api/media/:id` - Delete media (owner/admin)

### Admin Only
- `GET /api/users` - List all users
- `GET /api/users/:id` - Get user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user
- `POST /api/users/:id/roles` - Assign role
- `DELETE /api/users/:id/roles` - Remove role

## Important Notes

- **Password Security**: Passwords are hashed with bcrypt before storage; never stored in plaintext
- **JWT Secret**: Must be strong and kept secure; generate with `openssl rand -base64 32`
- **Database Migrations**: Auto-migration runs on startup; use caution in production
- **CORS**: Currently allows all origins (`*`); restrict in production
- **Soft Deletes**: Deleted records remain in DB but are excluded from queries
- **Authorization Header**: Use format `Authorization: Bearer <token>`
- **File Storage**: Currently local filesystem; consider cloud storage for production

## Testing

See `TESTING.md` for comprehensive curl examples and test scenarios.

To promote a user to admin:
```sql
-- Get user and role IDs first
INSERT INTO user_roles (user_id, role_id) 
VALUES ((SELECT id FROM users WHERE email='user@example.com'), 
        (SELECT id FROM roles WHERE name='admin'));
```

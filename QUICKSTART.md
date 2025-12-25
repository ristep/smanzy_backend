# Quick Start Guide

Get the User Management API running in minutes.

## Prerequisites

- Go 1.21+
- PostgreSQL 12+
- (Optional) Docker & Docker Compose

## Option 1: Local Development with Docker

### 1. Start PostgreSQL

```bash
docker-compose up -d
```

PostgreSQL will be available at `localhost:5432`
pgAdmin will be available at `http://localhost:5050` (admin@example.com / admin)

### 2. Install Dependencies

```bash
go mod download
```

### 3. Run the Application

```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

## Option 2: Using Existing PostgreSQL

### 1. Create Database

```bash
createdb smanzy_db
```

### 2. Update .env

Edit `.env` with your PostgreSQL credentials:

```env
DB_DSN=postgres://user:password@localhost:5432/smanzy_postgres?sslmode=disable
JWT_SECRET=your-super-secret-key
SERVER_PORT=8080
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run the Application

```bash
go run cmd/api/main.go
```

## Test the API

### 1. Register a User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123",
    "name": "John Doe"
  }'
```

Response will include `access_token` and `refresh_token`.

### 2. Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

### 3. Access Protected Route

Replace `<ACCESS_TOKEN>` with token from login response:

```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

### 4. Upload Media

```bash
curl -X POST http://localhost:8080/api/media \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -F "file=@/path/to/image.jpg"
```

### 5. Create Album

```bash
curl -X POST http://localhost:8080/api/albums \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Album",
    "description": "Album description"
  }'
```

### 6. Add Media to Album

```bash
curl -X POST http://localhost:8080/api/albums/1/media \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "media_id": 1
  }'
```

### 7. Get User Albums

```bash
curl -X GET http://localhost:8080/api/albums \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

## Create Admin User

Manually add admin role to a user via database or API:

### Via API (requires admin token):

```bash
# First create a user and get their ID
curl -X POST http://localhost:8080/api/users/<USER_ID>/roles \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "role_name": "admin"
  }'
```

### Via Database:

```sql
-- Connect to the database
psql -U postgres -d smanzy_db

-- Find a user
SELECT * FROM users;

-- Get the admin role ID
SELECT * FROM roles WHERE name = 'admin';

-- Create association
INSERT INTO user_roles (user_id, role_id) VALUES (1, 2);
```

## Use Development Tools

### Makefile Commands

```bash
make help              # Show all available commands
make build             # Build the application
make run               # Run the application
make test              # Run tests
make fmt               # Format code
make lint              # Run linter (requires golangci-lint)
make clean             # Clean build artifacts
make install-tools     # Install development tools
```

### Development with Hot Reload

```bash
# Install air (hot reload tool)
go install github.com/air-verse/air@latest

# Run with auto-reload
make dev
```

## Troubleshooting

### Database Connection Error

**Problem**: `Failed to connect to database`

**Solution**: 
- Check PostgreSQL is running: `psql -U postgres -d postgres -c "SELECT 1;"`
- Verify `DB_DSN` in `.env` is correct
- Check database exists: `psql -U postgres -l | grep smanzy_db`

### JWT Secret Error

**Problem**: `JWT_SECRET environment variable is required`

**Solution**: 
- Create `.env` file with `JWT_SECRET=<your-secret-key>`
- Or set environment variable: `export JWT_SECRET=your-secret-key`

### Port Already in Use

**Problem**: `bind: address already in use`

**Solution**:
- Change `SERVER_PORT` in `.env` to an available port
- Or kill process using port 8080: `lsof -i :8080 | grep LISTEN | awk '{print $2}' | xargs kill -9`

### Migration Errors

**Problem**: `Failed to auto-migrate models`

**Solution**:
- Drop and recreate database: `dropdb smanzy_db && createdb smanzy_db`
- Check PostgreSQL version is 12+

## Next Steps

1. Review [README.md](README.md) for complete API documentation
2. Explore the codebase structure in each directory
3. Run tests: `go test ./...`
4. Deploy to production following production guide in README

## Common cURL Examples

### Register
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "StrongPassword123",
    "name": "Jane Smith"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "StrongPassword123"
  }'
```

### Get Profile
```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>"
```

### Get All Users (Admin Only)
```bash
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer <ADMIN_ACCESS_TOKEN>"
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<YOUR_REFRESH_TOKEN>"
  }'
```

### Create Album
```bash
curl -X POST http://localhost:8080/api/albums \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Album",
    "description": "Album description"
  }'
```

### Get All User Albums
```bash
curl -X GET http://localhost:8080/api/albums \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>"
```

### Get Specific Album
```bash
curl -X GET http://localhost:8080/api/albums/1 \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>"
```

### Update Album
```bash
curl -X PUT http://localhost:8080/api/albums/1 \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title",
    "description": "Updated description"
  }'
```

### Upload Media File
```bash
curl -X POST http://localhost:8080/api/media \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -F "file=@/path/to/image.jpg"
```

### Add Media to Album
```bash
curl -X POST http://localhost:8080/api/albums/1/media \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "media_id": 1
  }'
```

### Remove Media from Album
```bash
curl -X DELETE http://localhost:8080/api/albums/1/media \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "media_id": 1
  }'
```

### Delete Album
```bash
curl -X DELETE http://localhost:8080/api/albums/1 \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>"
```

## Support

For detailed API documentation and advanced features, see [README.md](README.md)

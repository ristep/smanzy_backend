# Smanzy API with JWT Authentication

A production-ready REST API in Go featuring secure user management, role-based access control (RBAC), JWT authentication, and media file management.

## Tech Stack

- **Web Framework**: [Gin](https://github.com/gin-gonic/gin) - High-performance HTTP web framework
- **JWT Library**: [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) - JWT authentication
- **ORM**: [GORM](https://gorm.io/) - Object-relational mapping with auto-migrations
- **Password Hashing**: [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - Secure password storage
- **Database**: PostgreSQL (configurable via connection string)
- **Environment**: [godotenv](https://github.com/joho/godotenv) - Environment variable management

## Project Structure

```text
smanzy_backend/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── models/
│   │   ├── user.go                 # User and Role data models
│   │   ├── media.go                # Media data model
│   │   └── album.go                # Album data model with many-to-many media relationship
│   ├── handlers/
│   │   ├── auth.go                 # HTTP handlers for auth and user management
│   │   ├── media.go                # HTTP handlers for media management
│   │   └── album.go                # HTTP handlers for album management
│   ├── services/
│   │   └── album.go                # Business logic for album operations
│   ├── middleware/
│   │   ├── auth.go                 # JWT and RBAC middleware
│   │   └── cors.go                 # CORS configuration
│   └── auth/
│       └── jwt.go                  # JWT token generation and validation
├── uploads/                        # Local storage for media files
├── go.mod                           # Go module dependencies
├── go.sum                           # Go module checksums
├── Makefile                         # Common development tasks
├── .env.example                     # Example environment variables
└── README.md                        # This file
```

## Setup Instructions

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Git

### 1. Initialize the Project

```bash
cd smanzy_backend
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
# OR use Makefile
make run
```

The server will start on `http://localhost:8080`

### Optional: pgAdmin

You can run pgAdmin as a Docker container (this repository's `docker-compose.yml` includes a `pgadmin` service):

1. Start the containers:

```bash
docker-compose up -d
```

2. Visit pgAdmin at `http://localhost:5050` with the credentials specified in `docker-compose.yml`.

## API Endpoints

### Health Check

```http
GET /health
Response: {"status": "ok"}
```

### Public Endpoints

#### Register a New User

```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "John Doe",
  "tel": "+123456789",
  "age": 25,
  "gender": "male",
  "address": "123 Main St",
  "city": "Metropolis",
  "country": "USA"
}
```

#### Login

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

#### Public Media Listing

```http
GET /api/media?limit=100&offset=0
```

#### Serving Files (Development)

```http
GET /api/media/files/:name
```

### Protected Endpoints (Requires JWT)

#### Get User Profile

```http
GET /api/profile
```

#### Upload Media

```http
POST /api/media
Content-Type: multipart/form-data
Body: file (binary)
```

#### Update Media Metadata

```http
PUT /api/media/:id
Content-Type: application/json
{
  "filename": "new_name.jpg"
}
```

#### Delete Media

```http
DELETE /api/media/:id
```

### Album Management Endpoints (Requires JWT)

#### Create a New Album

```http
POST /api/albums
Content-Type: application/json

{
  "title": "My Vacation",
  "description": "Summer 2025 photos"
}
```

#### Get All User Albums

```http
GET /api/albums
```

#### Get Specific Album with Media

```http
GET /api/albums/:id
```

#### Update Album Details

```http
PUT /api/albums/:id
Content-Type: application/json

{
  "title": "Updated Title",
  "description": "Updated description"
}
```

#### Add Media to Album

```http
POST /api/albums/:id/media
Content-Type: application/json

{
  "media_id": 5
}
```

#### Remove Media from Album

```http
DELETE /api/albums/:id/media
Content-Type: application/json

{
  "media_id": 5
}
```

#### Delete Album (Soft Delete)

```http
DELETE /api/albums/:id
```

### Admin-Only Endpoints

- `GET /api/users` - List all users
- `GET /api/users/:id` - Get specific user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user
- `POST /api/users/:id/roles` - Assign role
- `DELETE /api/users/:id/roles` - Remove role

## Development

Use the included `Makefile` for common tasks:

- `make run`: Run the API
- `make dev`: Run with hot-reloading (requires `air`)
- `make build`: Build the binary
- `make test`: Run tests
- `make fmt`: Format code

## License

MIT License

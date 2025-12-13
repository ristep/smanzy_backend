# API Testing Guide

This file contains example requests for testing the User Management API.

## Setup

1. Start the application: `go run cmd/api/main.go`
2. The API will be available at `http://localhost:8080`
3. Store tokens from responses to use in subsequent requests

## Testing with cURL

### 1. Health Check

```bash
curl -X GET http://localhost:8080/health
```

Expected response:
```json
{"status": "ok"}
```

### 2. Register New User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "mano.bano@example.com",
    "password": "SecurePassword123",
    "name": "John Doe",
    "tel": "+1234567890",
    "age": 30,
    "address": "123 Main St",
    "city": "Metropolis",
    "country": "Utopia"
  }'
```

Save the `access_token` and `refresh_token` from the response.

Expected response (201 Created):
```json
{
  "data": {
    "user": {
      "id": 1,
      "email": "john.doe@example.com",
      "name": "John Doe",
      "email": "john.doe@example.com",
      "name": "John Doe",
      "tel": "+1234567890",
      "age": 30,
      "address": "123 Main St",
      "city": "Metropolis",
      "country": "Utopia",
      "email_verified": false,
      "roles": [
        {
          "id": 1,
          "name": "user"
        }
      ],
      "created_at": 1702324800000,
      "updated_at": 1702324800000
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 3. Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePassword123"
  }'
```

Expected response (200 OK):
```json
{
  "data": {
    "user": {...},
    "access_token": "...",
    "refresh_token": "..."
  }
}
```

### 4. Get User Profile (Protected)

Replace `YOUR_ACCESS_TOKEN` with the token from login:

```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

Expected response (200 OK):
```json
{
  "data": {
    "id": 1,
    "email": "john.doe@example.com",
    "name": "John Doe",
    "roles": [{"id": 1, "name": "user"}],
    "created_at": 1702324800000,
    "updated_at": 1702324800000
  }
}
```

### 5. Refresh Token

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

Expected response (200 OK):
```json
{
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  }
}
```

## Testing Admin Routes

To test admin-only routes, you need to assign the admin role to a user.

### Promote User to Admin

First, create another user or use the database:

#### Via Database

```sql
-- Connect to database
psql -U postgres -d smanzy_db

-- Find the user to promote
SELECT id FROM users WHERE email = 'john.doe@example.com';
-- Result: 1

-- Get the admin role ID
SELECT id FROM roles WHERE name = 'admin';
-- Result: 2

-- Create the association
INSERT INTO user_roles (user_id, role_id) VALUES (1, 2);
```

#### Via API

(Requires an existing admin token)

```bash
curl -X POST http://localhost:8080/api/users/1/roles \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_name": "admin"
  }'
```

### Admin Routes (After Promotion)

#### Get All Users

```bash
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

Expected response (200 OK):
```json
{
  "data": [
    {
      "id": 1,
      "email": "john.doe@example.com",
      "name": "John Doe",
      "roles": [
        {"id": 1, "name": "user"},
        {"id": 2, "name": "admin"}
      ],
      "created_at": 1702324800000,
      "updated_at": 1702324800000
    }
  ]
}
```

#### Get Specific User

```bash
curl -X GET http://localhost:8080/api/users/1 \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

#### Update User

```bash
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Smith",
    "tel": "+1987654321",
    "city": "Gotham"
  }'
```

#### Assign Role to User

```bash
curl -X POST http://localhost:8080/api/users/2/roles \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_name": "moderator"
  }'
```

#### Remove Role from User

```bash
curl -X DELETE http://localhost:8080/api/users/2/roles \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_name": "moderator"
  }'
```

#### Delete User

```bash
curl -X DELETE http://localhost:8080/api/users/2 \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

Expected response (200 OK):
```json
{
  "data": {
    "message": "User deleted successfully"
  }
}
```

## Testing with Postman

Import the following collection JSON into Postman:

```json
{
  "info": {
    "name": "User Management API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Register",
      "request": {
        "method": "POST",
        "url": "http://localhost:8080/api/auth/register",
        "header": [
          {"key": "Content-Type", "value": "application/json"}
        ],
        "body": {
          "mode": "raw",
          "raw": "{\"email\": \"user@example.com\", \"password\": \"Password123\", \"name\": \"User\"}"
        }
      }
    },
    {
      "name": "Login",
      "request": {
        "method": "POST",
        "url": "http://localhost:8080/api/auth/login",
        "header": [
          {"key": "Content-Type", "value": "application/json"}
        ],
        "body": {
          "mode": "raw",
          "raw": "{\"email\": \"user@example.com\", \"password\": \"Password123\"}"
        }
      }
    },
    {
      "name": "Get Profile",
      "request": {
        "method": "GET",
        "url": "http://localhost:8080/api/profile",
        "header": [
          {"key": "Authorization", "value": "Bearer {{access_token}}"}
        ]
      }
    },
    {
      "name": "Get All Users (Admin)",
      "request": {
        "method": "GET",
        "url": "http://localhost:8080/api/users",
        "header": [
          {"key": "Authorization", "value": "Bearer {{access_token}}"}
        ]
      }
    }
  ]
}
```

## Error Responses

### 400 Bad Request

```json
{
  "error": "Invalid input"
}
```

### 401 Unauthorized

```json
{
  "error": "Invalid or expired token"
}
```

### 403 Forbidden

```json
{
  "error": "Insufficient permissions"
}
```

### 404 Not Found

```json
{
  "error": "User not found"
}
```

### 500 Internal Server Error

```json
{
  "error": "Database error"
}
```

## Test Scenarios

### Scenario 1: Complete User Workflow

1. Register a new user
2. Login with that user
3. Get user profile
4. Refresh the token
5. Use new token to access profile again

### Scenario 2: Admin User Management

1. Create two users (User A and User B)
2. Promote User A to admin
3. As User A (admin), retrieve all users
4. As User A (admin), update User B
5. As User A (admin), assign role to User B
6. As User A (admin), remove role from User B
7. As User A (admin), delete User B

### Scenario 3: Authorization Checks

1. Create a regular user
2. Try to access GET /api/users without admin role (should fail with 403)
3. Promote to admin
4. Try again (should succeed)

### Scenario 4: Token Expiration (Simulated)

1. Login and get tokens
2. Decode JWT and note expiration time
3. Wait for access token to expire (or manually expire)
4. Try to access protected route (should fail with 401)
5. Use refresh token to get new access token
6. Try again (should succeed)

## Performance Testing

For load testing, use tools like:

- Apache Bench: `ab -n 1000 -c 10 -H "Authorization: Bearer TOKEN" http://localhost:8080/api/profile`
- Wrk: `wrk -t4 -c100 -d30s -H "Authorization: Bearer TOKEN" http://localhost:8080/api/profile`
- LoadRunner or JMeter for more advanced scenarios

## Security Testing Checklist

- [ ] Verify JWT secret is not logged
- [ ] Verify passwords are hashed (check DB)
- [ ] Verify password field is not in JSON responses
- [ ] Test with expired tokens
- [ ] Test with invalid signatures
- [ ] Test with wrong audience/issuer
- [ ] Test CORS headers
- [ ] Test SQL injection in inputs
- [ ] Test XSS in user inputs
- [ ] Verify role-based access enforcement

package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/ristep/smanzy_backend/internal/auth"
	"github.com/ristep/smanzy_backend/internal/models"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	db         *gorm.DB
	jwtService *auth.JWTService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtService: jwtService,
	}
}

// RegisterRequest represents the JSON payload for registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2"`
	Tel      string `json:"tel"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Country  string `json:"country"`
}

// LoginRequest represents the JSON payload for login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest represents the JSON payload for refresh token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error string `json:"error"`
}

// RegisterHandler handles user registration
func (ah *AuthHandler) RegisterHandler(c *gin.Context) {
	var req RegisterRequest

	// Validate JSON input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := ah.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "User already exists"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to process password"})
		return
	}

	// Get or create the default "user" role
	var userRole models.Role
	if err := ah.db.FirstOrCreate(&userRole, models.Role{Name: "user"}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Create the new user
	newUser := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		Tel:      req.Tel,
		Age:      req.Age,
		Gender:   req.Gender,
		Address:  req.Address,
		City:     req.City,
		Country:  req.Country,
		Roles:    []models.Role{userRole},
	}

	if err := ah.db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
		return
	}

	// Load the user with roles
	if err := ah.db.Preload("Roles").First(&newUser, newUser.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve user"})
		return
	}

	// Generate tokens
	tokenPair, err := ah.jwtService.GenerateTokenPair(&newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{Data: map[string]interface{}{
		"user":          newUser,
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	}})
}

// LoginHandler handles user login
func (ah *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest

	// Validate JSON input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	// Find user by email
	var user models.User
	if err := ah.db.Preload("Roles").Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		return
	}

	// Generate tokens
	tokenPair, err := ah.jwtService.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"user":          user,
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	}})
}

// RefreshHandler handles token refresh
func (ah *AuthHandler) RefreshHandler(c *gin.Context) {
	var req RefreshRequest

	// Validate JSON input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	// Validate the refresh token
	claims, err := ah.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	// Fetch the user from the database
	var user models.User
	if err := ah.db.Preload("Roles").First(&user, claims.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Generate a new token pair
	tokenPair, err := ah.jwtService.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	}})
}

// ProfileHandler returns the current user's profile
func (ah *AuthHandler) ProfileHandler(c *gin.Context) {
	// Get user from context (set by middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	userObj, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Invalid user data"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: userObj})
}

// UpdateProfileHandler updates the current user's profile
func (ah *AuthHandler) UpdateProfileHandler(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	userObj := user.(*models.User)

	// Update fields
	if req.Name != "" {
		userObj.Name = req.Name
	}
	if req.Tel != "" {
		userObj.Tel = req.Tel
	}
	if req.Age != 0 {
		userObj.Age = req.Age
	}
	if req.Address != "" {
		userObj.Address = req.Address
	}
	if req.City != "" {
		userObj.City = req.City
	}
	if req.Country != "" {
		userObj.Country = req.Country
	}
	if req.Gender != "" {
		userObj.Gender = req.Gender
	}

	if err := ah.db.Save(userObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: userObj})
}

// DeleteProfileHandler deletes the current user's profile
func (ah *AuthHandler) DeleteProfileHandler(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	userObj := user.(*models.User)

	// In GORM, when deleting an object that has Many-to-Many relationships (like Roles),
	// we should clear the associations first or ensure the join table is handled.
	// For a full delete of the user, we also need to be careful with the Soft Delete.

	// 1. Clear roles association
	if err := ah.db.Model(userObj).Association("Roles").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to clear roles"})
		return
	}

	// 2. Delete the user (this will be a soft delete because of DeletedAt field)
	if err := ah.db.Delete(userObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete profile"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: map[string]string{"message": "Profile deleted successfully"}})
}

// UserHandler represents handlers for user management
type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// GetAllUsersHandler returns all users (admin only)
func (uh *UserHandler) GetAllUsersHandler(c *gin.Context) {
	var users []models.User

	if err := uh.db.Preload("Roles").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: users})
}

// GetUserByIDHandler returns a specific user by ID (admin only)
func (uh *UserHandler) GetUserByIDHandler(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := uh.db.Preload("Roles").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: user})
}

// UpdateUserRequest represents the JSON payload for user updates
type UpdateUserRequest struct {
	Name    string `json:"name" binding:"omitempty,min=2"`
	Tel     string `json:"tel"`
	Age     int    `json:"age"`
	Address string `json:"address"`
	City    string `json:"city"`
	Country string `json:"country"`
	Gender  string `json:"gender"`
}

// UpdateUserHandler updates a user (user can update self, admin can update anyone)
func (uh *UserHandler) UpdateUserHandler(c *gin.Context) {
	userID := c.Param("id")
	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	// Get current user from context
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	currentUserObj := currentUser.(*models.User)

	// Check if user is trying to update someone else (must be admin)
	if userID != strconv.FormatUint(uint64(currentUserObj.ID), 10) {
		// Check if current user is admin
		isAdmin := false
		for _, role := range currentUserObj.Roles {
			if role.Name == "admin" {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
			return
		}
	}

	var user models.User
	if err := uh.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Tel != "" {
		user.Tel = req.Tel
	}
	if req.Age != 0 {
		user.Age = req.Age
	}
	if req.Address != "" {
		user.Address = req.Address
	}
	if req.City != "" {
		user.City = req.City
	}
	if req.Country != "" {
		user.Country = req.Country
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}

	if err := uh.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update user"})
		return
	}

	// Reload with roles
	uh.db.Preload("Roles").First(&user, userID)

	c.JSON(http.StatusOK, SuccessResponse{Data: user})
}

// DeleteUserHandler deletes a user (admin only)
func (uh *UserHandler) DeleteUserHandler(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := uh.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	if err := uh.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: map[string]string{"message": "User deleted successfully"}})
}

// AssignRoleRequest represents the JSON payload for assigning roles
type AssignRoleRequest struct {
	RoleName string `json:"role_name" binding:"required"`
}

// AssignRoleHandler assigns a role to a user (admin only)
func (uh *UserHandler) AssignRoleHandler(c *gin.Context) {
	userID := c.Param("id")
	var req AssignRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	// Normalize role name
	roleName := strings.ToLower(strings.TrimSpace(req.RoleName))

	var user models.User
	if err := uh.db.Preload("Roles").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Find or create the role
	var role models.Role
	if err := uh.db.FirstOrCreate(&role, models.Role{Name: roleName}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Check if user already has this role
	for _, r := range user.Roles {
		if r.ID == role.ID {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "User already has this role"})
			return
		}
	}

	// Assign the role
	if err := uh.db.Model(&user).Association("Roles").Append(&role); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to assign role"})
		return
	}

	// Reload user with roles
	uh.db.Preload("Roles").First(&user, userID)

	c.JSON(http.StatusOK, SuccessResponse{Data: user})
}

// RemoveRoleRequest represents the JSON payload for removing roles
type RemoveRoleRequest struct {
	RoleName string `json:"role_name" binding:"required"`
}

// RemoveRoleHandler removes a role from a user (admin only)
func (uh *UserHandler) RemoveRoleHandler(c *gin.Context) {
	userID := c.Param("id")
	var req RemoveRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	roleName := strings.ToLower(strings.TrimSpace(req.RoleName))

	var user models.User
	if err := uh.db.Preload("Roles").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	var roleToRemove *models.Role
	for i := range user.Roles {
		if user.Roles[i].Name == roleName {
			roleToRemove = &user.Roles[i]
			break
		}
	}

	if roleToRemove == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "User doesn't have this role"})
		return
	}

	// Remove the role
	if err := uh.db.Model(&user).Association("Roles").Delete(roleToRemove); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove role"})
		return
	}

	// Reload user with roles
	uh.db.Preload("Roles").First(&user, userID)

	c.JSON(http.StatusOK, SuccessResponse{Data: user})
}

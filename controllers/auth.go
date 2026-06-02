// Package controllers handles all incoming HTTP requests and sends responses.
// auth.go handles user registration, login, and health check endpoints.
package controllers

import (
	"encoding/json"
	"regexp"
	"strings"

	"expense-tracker-api/models"

	"github.com/beego/beego/v2/core/logs"
)

// AuthController handles authentication-related endpoints
// including user registration, login, and health check.
type AuthController struct {
	BaseController
}

// registerInput defines the expected JSON body for the register endpoint.
type registerInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginInput defines the expected JSON body for the login endpoint.
type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse defines the data returned on successful login.
type loginResponse struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// HealthCheck handles GET /api/v1/health
// @Summary Check if the server is running
// @Tags Health
// @Produce json
// @Success 200 {object} ResponseData
// @Router /api/v1/health [get]
func (c *AuthController) HealthCheck() {
	logs.Info("Health check endpoint called")
	c.SendSuccess("Server is running", nil)
}

// Register handles POST /api/v1/auth/register
// @Summary Register a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body registerInput true "User registration details"
// @Success 201 {object} ResponseData
// @Failure 400 {object} ResponseData
// @Failure 409 {object} ResponseData
// @Router /api/v1/auth/register [post]
func (c *AuthController) Register() {
	logs.Info("Register endpoint called")

	var input registerInput

	// Parse JSON request body
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Register: invalid JSON body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	// Trim whitespace from all fields
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.Password = strings.TrimSpace(input.Password)

	// Validate: name is required
	if input.Name == "" {
		c.SendError(400, "Name is required")
		return
	}

	// Validate: email is required and must be valid format
	if input.Email == "" {
		c.SendError(400, "Email is required")
		return
	}
	if !isValidEmail(input.Email) {
		c.SendError(400, "Invalid email format")
		return
	}

	// Validate: password is required and minimum 6 characters
	if input.Password == "" {
		c.SendError(400, "Password is required")
		return
	}
	if len(input.Password) < 6 {
		c.SendError(400, "Password must be at least 6 characters")
		return
	}

	// Check if email already exists
	existingUser, _ := models.GetUserByEmail(input.Email)
	if existingUser != nil {
		logs.Warn("Register: email already exists:", input.Email)
		c.SendError(409, "Email already exists")
		return
	}

	// Build new user and save to CSV
	newUser := &models.User{
		ID:       models.GetNextUserID(),
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	}

	if err := models.CreateUser(newUser); err != nil {
		logs.Error("Register: failed to create user:", err)
		c.SendError(500, "Failed to create user")
		return
	}

	logs.Info("Register: user created successfully, ID:", newUser.ID)
	c.SendCreated("User registered successfully", nil)
}

// Login handles POST /api/v1/auth/login
// @Summary Login with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body loginInput true "Login credentials"
// @Success 200 {object} ResponseData
// @Failure 400 {object} ResponseData
// @Failure 401 {object} ResponseData
// @Router /api/v1/auth/login [post]
func (c *AuthController) Login() {
	logs.Info("Login endpoint called")

	var input loginInput

	// Parse JSON request body
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Login: invalid JSON body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	// Trim whitespace
	input.Email = strings.TrimSpace(input.Email)
	input.Password = strings.TrimSpace(input.Password)

	// Validate: email and password are required
	if input.Email == "" || input.Password == "" {
		c.SendError(400, "Email and password are required")
		return
	}

	// Find user by email
	user, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Warn("Login: user not found for email:", input.Email)
		c.SendError(401, "Invalid email or password")
		return
	}

	// Check password
	if user.Password != input.Password {
		logs.Warn("Login: wrong password for email:", input.Email)
		c.SendError(401, "Invalid email or password")
		return
	}

	logs.Info("Login: successful for user ID:", user.ID)
	c.SendSuccess("Login successful", loginResponse{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email,
	})
}

// isValidEmail checks if the given string is a valid email format.
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

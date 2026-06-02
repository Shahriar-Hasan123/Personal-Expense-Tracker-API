package controllers

import (
	"encoding/json"
	"expense-tracker-api/models"
	"regexp"
	"strings"

	"github.com/beego/beego/v2/core/logs"
)

// AuthController handles user registration and login.
type AuthController struct {
	BaseController
}

// registerInput defines the expected JSON body for registration.
type registerInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginInput defines the expected JSON body for login.
type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// isValidEmail checks if the email has a valid format.
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// Register handles POST /api/v1/auth/register
// @Title Register
// @Summary Register a new user
// @Description Creates a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param body body registerInput true "User registration details"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 409 {object} map[string]interface{} "Email already exists"
// @Router /auth/register [post]
func (c *AuthController) Register() {
	var input registerInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.respondBadRequest("Invalid request body")
		return
	}

	// Validation
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)

	if input.Name == "" {
		c.respondBadRequest("Name is required")
		return
	}
	if input.Email == "" || !isValidEmail(input.Email) {
		c.respondBadRequest("Valid email is required")
		return
	}
	if len(input.Password) < 6 {
		c.respondBadRequest("Password must be at least 6 characters")
		return
	}

	// Check for duplicate email
	existing, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Error("Register: GetUserByEmail error: %v", err)
		c.respondInternalError("Internal server error")
		return
	}
	if existing != nil {
		c.respondConflict("Email already exists")
		return
	}

	// Create user
	user := &models.User{
		ID:       models.GetNextID(),
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	}
	if err := models.CreateUser(user); err != nil {
		logs.Error("Register: CreateUser error: %v", err)
		c.respondInternalError("Failed to create user")
		return
	}

	logs.Info("New user registered: %s", input.Email)
	c.respondCreated("User registered successfully", nil)
}

// Login handles POST /api/v1/auth/login
// @Title Login
// @Summary Login with email and password
// @Description Authenticates a user and returns user info
// @Tags auth
// @Accept json
// @Produce json
// @Param body body loginInput true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]interface{} "Missing fields"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Router /auth/login [post]
func (c *AuthController) Login() {
	var input loginInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.respondBadRequest("Invalid request body")
		return
	}

	input.Email = strings.TrimSpace(input.Email)

	if input.Email == "" || input.Password == "" {
		c.respondBadRequest("Email and password are required")
		return
	}

	user, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Error("Login: GetUserByEmail error: %v", err)
		c.respondInternalError("Internal server error")
		return
	}

	if user == nil || user.Password != input.Password {
		c.respondUnauthorized("Invalid email or password")
		return
	}

	logs.Info("User logged in: %s", input.Email)
	c.respondSuccess("Login successful", map[string]interface{}{
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
	})
}

package controllers

import (
	"bytes"
	"encoding/json"
	"expense-tracker-api/models"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
	beegoContext "github.com/beego/beego/v2/server/web/context"
)

// setupAuthTestEnv wires temp CSVs and returns cleanup.
func setupAuthTestEnv(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	beego.AppConfig.Set("userscsv", filepath.Join(dir, "users.csv"))
	beego.AppConfig.Set("expensescsv", filepath.Join(dir, "expenses.csv"))
	beego.BConfig.WebConfig.AutoRender = false
	return func() { os.RemoveAll(dir) }
}

// execAuth wires an AuthController, calls the named method, returns recorder.
func execAuth(t *testing.T, method, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctx := beegoContext.NewContext()
	ctx.Reset(rr, req)
	ctx.Input.RequestBody = []byte(body)

	c := &AuthController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	switch method {
	case "Register":
		c.Register()
	case "Login":
		c.Login()
	}
	return rr
}

// parseBody decodes the response body into a generic map.
func parseBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v — body: %s", err, rr.Body.String())
	}
	return result
}

func TestRegister(t *testing.T) {
	defer setupAuthTestEnv(t)()

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "valid registration",
			body:       `{"name":"John Doe","email":"john@example.com","password":"secret123"}`,
			wantStatus: 201,
			wantMsg:    "User registered successfully",
		},
		{
			name:       "missing name",
			body:       `{"email":"john2@example.com","password":"secret123"}`,
			wantStatus: 400,
			wantMsg:    "Name is required",
		},
		{
			name:       "missing email",
			body:       `{"name":"John","password":"secret123"}`,
			wantStatus: 400,
			wantMsg:    "Valid email is required",
		},
		{
			name:       "invalid email format",
			body:       `{"name":"John","email":"not-an-email","password":"secret123"}`,
			wantStatus: 400,
			wantMsg:    "Valid email is required",
		},
		{
			name:       "password too short",
			body:       `{"name":"John","email":"john3@example.com","password":"abc"}`,
			wantStatus: 400,
			wantMsg:    "Password must be at least 6 characters",
		},
		{
			name:       "duplicate email",
			body:       `{"name":"John Doe","email":"john@example.com","password":"secret123"}`,
			wantStatus: 409,
			wantMsg:    "Email already exists",
		},
		{
			name:       "malformed JSON",
			body:       `{invalid}`,
			wantStatus: 400,
			wantMsg:    "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := execAuth(t, "Register", tt.body)
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			body := parseBody(t, rr)
			if body["message"] != tt.wantMsg {
				t.Errorf("message = %q, want %q", body["message"], tt.wantMsg)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	defer setupAuthTestEnv(t)()

	// Seed a user to log in with
	_ = models.CreateUser(&models.User{
		ID: 1, Name: "Jane", Email: "jane@example.com", Password: "password123",
	})

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "valid login",
			body:       `{"email":"jane@example.com","password":"password123"}`,
			wantStatus: 200,
			wantMsg:    "Login successful",
		},
		{
			name:       "wrong password",
			body:       `{"email":"jane@example.com","password":"wrongpass"}`,
			wantStatus: 401,
			wantMsg:    "Invalid email or password",
		},
		{
			name:       "email not found",
			body:       `{"email":"ghost@example.com","password":"password123"}`,
			wantStatus: 401,
			wantMsg:    "Invalid email or password",
		},
		{
			name:       "missing email",
			body:       `{"password":"password123"}`,
			wantStatus: 400,
			wantMsg:    "Email and password are required",
		},
		{
			name:       "missing password",
			body:       `{"email":"jane@example.com"}`,
			wantStatus: 400,
			wantMsg:    "Email and password are required",
		},
		{
			name:       "malformed JSON",
			body:       `{bad}`,
			wantStatus: 400,
			wantMsg:    "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := execAuth(t, "Login", tt.body)
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			body := parseBody(t, rr)
			if body["message"] != tt.wantMsg {
				t.Errorf("message = %q, want %q", body["message"], tt.wantMsg)
			}
		})
	}
}

// TestIsValidEmail tests the unexported email validator directly.
func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		want  bool
	}{
		{"user@example.com", true},
		{"user.name+tag@sub.domain.org", true},
		{"not-an-email", false},
		{"missing@", false},
		{"@nodomain.com", false},
		{"", false},
		{"spaces in@email.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := isValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}
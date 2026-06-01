package models

import (
	"os"
	"path/filepath"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
)

// setupUserTestEnv points the model at a temp CSV and returns cleanup func.
func setupUserTestEnv(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	csv := filepath.Join(dir, "users.csv")
	beego.AppConfig.Set("userscsv", csv)
	return func() { os.Remove(csv) }
}

func TestGetAllUsers_Empty(t *testing.T) {
	defer setupUserTestEnv(t)()

	users, err := GetAllUsers()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			name: "valid user",
			user: &User{ID: 1, Name: "Alice", Email: "alice@example.com", Password: "pass123"},
		},
		{
			name: "second user gets different ID",
			user: &User{ID: 2, Name: "Bob", Email: "bob@example.com", Password: "pass456"},
		},
	}

	defer setupUserTestEnv(t)()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateUser(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify both users persisted
	users, err := GetAllUsers()
	if err != nil {
		t.Fatalf("GetAllUsers() error: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestGetUserByEmail(t *testing.T) {
	defer setupUserTestEnv(t)()

	// Seed a user
	_ = CreateUser(&User{ID: 1, Name: "Carol", Email: "carol@example.com", Password: "pass123"})

	tests := []struct {
		name      string
		email     string
		wantFound bool
	}{
		{
			name:      "existing email",
			email:     "carol@example.com",
			wantFound: true,
		},
		{
			name:      "non-existing email",
			email:     "nobody@example.com",
			wantFound: false,
		},
		{
			name:      "empty email",
			email:     "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := GetUserByEmail(tt.email)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if (user != nil) != tt.wantFound {
				t.Errorf("found = %v, wantFound = %v", user != nil, tt.wantFound)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	defer setupUserTestEnv(t)()

	_ = CreateUser(&User{ID: 1, Name: "Dave", Email: "dave@example.com", Password: "pass123"})
	_ = CreateUser(&User{ID: 2, Name: "Eve", Email: "eve@example.com", Password: "pass456"})

	tests := []struct {
		name      string
		id        int
		wantFound bool
		wantName  string
	}{
		{
			name:      "find user ID 1",
			id:        1,
			wantFound: true,
			wantName:  "Dave",
		},
		{
			name:      "find user ID 2",
			id:        2,
			wantFound: true,
			wantName:  "Eve",
		},
		{
			name:      "non-existing ID",
			id:        999,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := GetUserByID(tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if (user != nil) != tt.wantFound {
				t.Errorf("found = %v, wantFound = %v", user != nil, tt.wantFound)
			}
			if tt.wantFound && user.Name != tt.wantName {
				t.Errorf("name = %q, want %q", user.Name, tt.wantName)
			}
		})
	}
}

func TestGetNextUserID(t *testing.T) {
	defer setupUserTestEnv(t)()

	tests := []struct {
		name   string
		seed   []*User
		wantID int
	}{
		{
			name:   "empty CSV returns 1",
			seed:   nil,
			wantID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, u := range tt.seed {
				_ = CreateUser(u)
			}
			got := GetNextID()
			if got != tt.wantID {
				t.Errorf("GetNextUserID() = %d, want %d", got, tt.wantID)
			}
		})
	}

	// After seeding users, next ID must increment
	_ = CreateUser(&User{ID: 1, Name: "A", Email: "a@example.com", Password: "pass123"})
	_ = CreateUser(&User{ID: 2, Name: "B", Email: "b@example.com", Password: "pass123"})
	got := GetNextID()
	if got != 3 {
		t.Errorf("after 2 users, GetNextUserID() = %d, want 3", got)
	}
}

// TestEnsureFile tests that the data directory and file are auto-created.
func TestEnsureFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "test.csv")

	err := ensureFile(path)
	if err != nil {
		t.Fatalf("ensureFile() error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist at %s", path)
	}
}
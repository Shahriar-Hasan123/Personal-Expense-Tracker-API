package models

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

// User represents a registered user in the system.
type User struct {
	ID        int
	Name      string
	Email     string
	Password  string
	CreatedAt string
}

// usersFilePath returns the path to the users CSV file from config.
func usersFilePath() string {
	path, _ := beego.AppConfig.String("userscsv")
	if path == "" {
		return "data/users.csv"
	}
	return path
}

// ensureFile creates the file and its parent directory if they don't exist.
func ensureFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

// GetAllUsers reads all users from the CSV file.
func GetAllUsers() ([]User, error) {
	path := usersFilePath()
	if err := ensureFile(path); err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var users []User
	for i, record := range records {
		// Skip header row
		if i == 0 && record[0] == "id" {
			continue
		}
		id, _ := strconv.Atoi(record[0])
		users = append(users, User{
			ID:        id,
			Name:      record[1],
			Email:     record[2],
			Password:  record[3],
			CreatedAt: record[4],
		})
	}
	return users, nil
}

// GetUserByEmail finds a user by their email address.
func GetUserByEmail(email string) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, nil
}

// GetUserByID finds a user by their ID.
func GetUserByID(id int) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, nil
}

// GetNextID determines the next available user ID.
func GetNextID() int {
	users, err := GetAllUsers()
	if err != nil || len(users) == 0 {
		return 1
	}
	max := 0
	for _, u := range users {
		if u.ID > max {
			max = u.ID
		}
	}
	return max + 1
}

// CreateUser appends a new user to the CSV file.
func CreateUser(user *User) error {
	path := usersFilePath()
	if err := ensureFile(path); err != nil {
		return err
	}

	// Check if file is empty to write header
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	writeHeader := info.Size() == 0

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	if writeHeader {
		if err := writer.Write([]string{"id", "name", "email", "password", "created_at"}); err != nil {
			return errors.New("failed to write header")
		}
	}

	user.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	record := []string{
		strconv.Itoa(user.ID),
		user.Name,
		user.Email,
		user.Password,
		user.CreatedAt,
	}
	return writer.Write(record)
}

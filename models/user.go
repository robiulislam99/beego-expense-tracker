// Package models handles all data structures and CSV file operations.
// user.go defines the User model and all user-related CSV functions.
package models

import (
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

// User represents a registered user in the system.
// All user data is stored and read from a CSV file.
type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}

// getUsersCSVPath reads the CSV file path from app.conf.
// Falls back to a default path if config is not set.
func getUsersCSVPath() string {
	path, _ := beego.AppConfig.String("users_csv_path")
	if path == "" {
		return "data/users.csv"
	}
	return path
}

// GetAllUsers reads all users from the CSV file and returns them as a slice.
// Returns an empty slice if the file does not exist yet.
func GetAllUsers() ([]User, error) {
	filePath := getUsersCSVPath()

	// If file does not exist yet, return empty list
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []User{}, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		logs.Error("Failed to open users CSV:", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		logs.Error("Failed to read users CSV:", err)
		return nil, err
	}

	var users []User

	// Skip header row at index 0
	for i, record := range records {
		if i == 0 {
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			logs.Warn("Skipping user row with invalid ID:", record)
			continue
		}

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

// GetUserByEmail searches the users CSV for a user with the given email.
// Returns a pointer to the User if found, or an error if not found.
func GetUserByEmail(email string) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Email == email {
			return &user, nil
		}
	}

	return nil, errors.New("user not found")
}

// GetUserByID searches the users CSV for a user with the given ID.
// Returns a pointer to the User if found, or an error if not found.
func GetUserByID(id int) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.ID == id {
			return &user, nil
		}
	}

	return nil, errors.New("user not found")
}

// GetNextUserID determines the next available user ID
// by finding the maximum existing ID and adding 1.
func GetNextUserID() int {
	users, err := GetAllUsers()
	if err != nil || len(users) == 0 {
		return 1
	}

	maxID := 0
	for _, user := range users {
		if user.ID > maxID {
			maxID = user.ID
		}
	}

	return maxID + 1
}

// CreateUser appends a new user record to the CSV file.
// Creates the file with a header row if it does not exist yet.
func CreateUser(user *User) error {
	filePath := getUsersCSVPath()

	// Check if file already exists to avoid writing header twice
	fileExists := true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fileExists = false
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logs.Error("Failed to open users CSV for writing:", err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header only when creating the file for the first time
	if !fileExists {
		header := []string{"id", "name", "email", "password", "created_at"}
		if err := writer.Write(header); err != nil {
			logs.Error("Failed to write users CSV header:", err)
			return err
		}
	}

	// Set creation timestamp if not already provided
	if user.CreatedAt == "" {
		user.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	record := []string{
		strconv.Itoa(user.ID),
		user.Name,
		user.Email,
		user.Password,
		user.CreatedAt,
	}

	if err := writer.Write(record); err != nil {
		logs.Error("Failed to write user record to CSV:", err)
		return err
	}

	logs.Info("New user created with ID:", user.ID)
	return nil
}

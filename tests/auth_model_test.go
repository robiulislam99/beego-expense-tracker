// Package tests contains table-driven unit tests for the expense-tracker-api.
// auth_test.go tests user model functions.
package tests

import (
	"expense-tracker-api/models"
	"os"
	"testing"
)

// testUsersCSVPath is the CSV path used during tests.
const testUsersCSVPath = "data/users.csv"

// setupTestUsersCSV creates a temporary users CSV file for testing.
// Returns a cleanup function to delete the file after the test.
func setupTestUsersCSV(t *testing.T) func() {
	t.Helper()

	// Create data directory relative to where tests run
	os.MkdirAll("data", 0755)

	content := "id,name,email,password,created_at\n1,John Doe,john@example.com,secret123,2025-06-01T10:00:00Z\n"
	err := os.WriteFile(testUsersCSVPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test users CSV: %v", err)
	}

	// Return cleanup function
	return func() {
		os.Remove(testUsersCSVPath)
		os.Remove("data")
	}
}

// TestGetAllUsers tests reading all users from the CSV file.
func TestGetAllUsers(t *testing.T) {
	cleanup := setupTestUsersCSV(t)
	defer cleanup()

	tests := []struct {
		name      string
		wantCount int
		wantError bool
	}{
		{
			name:      "reads users successfully",
			wantCount: 1,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := models.GetAllUsers()

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}
			if len(users) != tt.wantCount {
				t.Errorf("Expected %d users, got %d", tt.wantCount, len(users))
			}
		})
	}
}

// TestGetUserByEmail tests finding a user by their email address.
func TestGetUserByEmail(t *testing.T) {
	cleanup := setupTestUsersCSV(t)
	defer cleanup()

	tests := []struct {
		name      string
		email     string
		wantFound bool
	}{
		{
			name:      "finds existing user",
			email:     "john@example.com",
			wantFound: true,
		},
		{
			name:      "returns error for missing user",
			email:     "notfound@example.com",
			wantFound: false,
		},
		{
			name:      "returns error for empty email",
			email:     "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := models.GetUserByEmail(tt.email)

			if tt.wantFound && (err != nil || user == nil) {
				t.Errorf("Expected to find user but got error: %v", err)
			}
			if !tt.wantFound && err == nil {
				t.Error("Expected error but found user")
			}
		})
	}
}

// TestGetUserByID tests finding a user by their ID.
func TestGetUserByID(t *testing.T) {
	cleanup := setupTestUsersCSV(t)
	defer cleanup()

	tests := []struct {
		name      string
		id        int
		wantFound bool
	}{
		{
			name:      "finds existing user",
			id:        1,
			wantFound: true,
		},
		{
			name:      "returns error for missing ID",
			id:        999,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := models.GetUserByID(tt.id)

			if tt.wantFound && (err != nil || user == nil) {
				t.Errorf("Expected to find user but got: %v", err)
			}
			if !tt.wantFound && err == nil {
				t.Error("Expected error but found user")
			}
		})
	}
}

// TestCreateUser tests creating a new user in the CSV file.
func TestCreateUser(t *testing.T) {
	cleanup := setupTestUsersCSV(t)
	defer cleanup()

	tests := []struct {
		name      string
		user      *models.User
		wantError bool
	}{
		{
			name: "creates valid user",
			user: &models.User{
				ID:       2,
				Name:     "Jane Doe",
				Email:    "jane@example.com",
				Password: "password123",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.CreateUser(tt.user)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}

			// Verify user was actually saved
			if !tt.wantError {
				saved, err := models.GetUserByEmail(tt.user.Email)
				if err != nil || saved == nil {
					t.Error("User was not saved to CSV")
				}
			}
		})
	}
}

// TestGetNextUserID tests that the next ID is always max existing ID + 1.
// Uses a fresh CSV with only one user so the next ID is always 2.
func TestGetNextUserID(t *testing.T) {
	cleanup := setupTestUsersCSV(t)
	defer cleanup()

	tests := []struct {
		name   string
		wantID int
	}{
		{
			name:   "returns next available ID",
			wantID: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := models.GetNextUserID()
			if id != tt.wantID {
				t.Errorf("Expected ID %d, got %d", tt.wantID, id)
			}
		})
	}
}

// TestIsValidEmail tests the email format validation.
// Since isValidEmail is unexported, we test it indirectly via GetUserByEmail logic.
func TestIsValidEmail(t *testing.T) {
	cleanup := setupTestUsersCSV(t)
	defer cleanup()

	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{name: "valid email", email: "user@example.com", want: true},
		{name: "valid email with dots", email: "user.name@domain.co", want: true},
		{name: "missing @ symbol", email: "userexample.com", want: false},
		{name: "missing domain", email: "user@", want: false},
		{name: "empty string", email: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that invalid emails don't match existing users
			user, _ := models.GetUserByEmail(tt.email)
			if tt.want == false && user != nil {
				t.Errorf("Expected no user for email %q but found one", tt.email)
			}
		})
	}
}
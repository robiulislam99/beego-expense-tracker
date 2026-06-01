// Package tests contains table-driven unit tests for the expense-tracker-api.
// expense_test.go tests expense model functions and validation logic.
package tests

import (
	"expense-tracker-api/models"
	"os"
	"testing"
)

// testExpensesCSVPath is the CSV path used during tests.
const testExpensesCSVPath = "data/expenses.csv"

// setupTestExpensesCSV creates a temporary expenses CSV file for testing.
// Returns a cleanup function to delete the file after the test.
func setupTestExpensesCSV(t *testing.T) func() {
	t.Helper()

	os.MkdirAll("data", 0755)

	content := "id,user_id,title,amount,category,note,expense_date,created_at\n" +
		"1,1,Lunch,350.50,Food,Team lunch,2025-06-10,2025-06-10T14:30:00Z\n" +
		"2,1,Bus fare,50.00,Transport,Morning commute,2025-06-11,2025-06-11T08:00:00Z\n" +
		"3,2,Groceries,800.00,Food,Weekly groceries,2025-06-12,2025-06-12T10:00:00Z\n"

	err := os.WriteFile(testExpensesCSVPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test expenses CSV: %v", err)
	}

	return func() {
		os.Remove(testExpensesCSVPath)
		os.Remove("data")
	}
}

// TestGetAllExpenses tests reading all expenses from the CSV file.
func TestGetAllExpenses(t *testing.T) {
	cleanup := setupTestExpensesCSV(t)
	defer cleanup()

	tests := []struct {
		name      string
		wantCount int
		wantError bool
	}{
		{
			name:      "reads all expenses successfully",
			wantCount: 3,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenses, err := models.GetAllExpenses()

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}
			if len(expenses) != tt.wantCount {
				t.Errorf("Expected %d expenses, got %d", tt.wantCount, len(expenses))
			}
		})
	}
}

// TestGetExpensesByUserID tests filtering expenses by user ID.
func TestGetExpensesByUserID(t *testing.T) {
	cleanup := setupTestExpensesCSV(t)
	defer cleanup()

	tests := []struct {
		name      string
		userID    int
		wantCount int
	}{
		{
			name:      "returns expenses for user 1",
			userID:    1,
			wantCount: 2,
		},
		{
			name:      "returns expenses for user 2",
			userID:    2,
			wantCount: 1,
		},
		{
			name:      "returns empty for unknown user",
			userID:    999,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenses, err := models.GetExpensesByUserID(tt.userID)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(expenses) != tt.wantCount {
				t.Errorf("Expected %d expenses, got %d", tt.wantCount, len(expenses))
			}
		})
	}
}

// TestGetExpenseByID tests finding a single expense by ID and user ID.
func TestGetExpenseByID(t *testing.T) {
	cleanup := setupTestExpensesCSV(t)
	defer cleanup()

	tests := []struct {
		name      string
		id        int
		userID    int
		wantFound bool
	}{
		{
			name:      "finds existing expense",
			id:        1,
			userID:    1,
			wantFound: true,
		},
		{
			name:      "returns error for wrong user",
			id:        1,
			userID:    2,
			wantFound: false,
		},
		{
			name:      "returns error for missing ID",
			id:        999,
			userID:    1,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense, err := models.GetExpenseByID(tt.id, tt.userID)

			if tt.wantFound && (err != nil || expense == nil) {
				t.Errorf("Expected to find expense but got: %v", err)
			}
			if !tt.wantFound && err == nil {
				t.Error("Expected error but found expense")
			}
		})
	}
}

// TestCreateExpense tests creating a new expense in the CSV file.
func TestCreateExpense(t *testing.T) {
	cleanup := setupTestExpensesCSV(t)
	defer cleanup()

	tests := []struct {
		name      string
		expense   *models.Expense
		wantError bool
	}{
		{
			name: "creates valid expense",
			expense: &models.Expense{
				ID:          4,
				UserID:      1,
				Title:       "Dinner",
				Amount:      500.00,
				Category:    "Food",
				Note:        "Family dinner",
				ExpenseDate: "2025-06-13",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.CreateExpense(tt.expense)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}

			// Verify expense was saved
			if !tt.wantError {
				saved, err := models.GetExpenseByID(tt.expense.ID, tt.expense.UserID)
				if err != nil || saved == nil {
					t.Error("Expense was not saved to CSV")
				}
			}
		})
	}
}

// TestUpdateExpense tests updating an existing expense in the CSV file.
func TestUpdateExpense(t *testing.T) {
	cleanup := setupTestExpensesCSV(t)
	defer cleanup()

	tests := []struct {
		name      string
		expense   *models.Expense
		wantError bool
	}{
		{
			name: "updates existing expense",
			expense: &models.Expense{
				ID:          1,
				UserID:      1,
				Title:       "Updated Lunch",
				Amount:      400.00,
				Category:    "Food",
				Note:        "Updated note",
				ExpenseDate: "2025-06-10",
				CreatedAt:   "2025-06-10T14:30:00Z",
			},
			wantError: false,
		},
		{
			name: "returns error for missing expense",
			expense: &models.Expense{
				ID:          999,
				UserID:      1,
				Title:       "Ghost",
				Amount:      100.00,
				Category:    "Food",
				ExpenseDate: "2025-06-10",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.UpdateExpense(tt.expense)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}
		})
	}
}

// TestDeleteExpense tests deleting an expense from the CSV file.
func TestDeleteExpense(t *testing.T) {
	cleanup := setupTestExpensesCSV(t)
	defer cleanup()

	tests := []struct {
		name      string
		id        int
		userID    int
		wantError bool
	}{
		{
			name:      "deletes existing expense",
			id:        1,
			userID:    1,
			wantError: false,
		},
		{
			name:      "returns error for missing expense",
			id:        999,
			userID:    1,
			wantError: true,
		},
		{
			name:      "returns error for wrong user",
			id:        2,
			userID:    99,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fresh CSV for each sub-test to avoid state sharing
			setupTestExpensesCSV(t)

			err := models.DeleteExpense(tt.id, tt.userID)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}
		})
	}
}

// TestIsValidCategory tests the category validation function.
func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		want     bool
	}{
		{name: "valid category Food", category: "Food", want: true},
		{name: "valid category Transport", category: "Transport", want: true},
		{name: "valid category Other", category: "Other", want: true},
		{name: "invalid category", category: "InvalidCat", want: false},
		{name: "empty string", category: "", want: false},
		{name: "lowercase food", category: "food", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.IsValidCategory(tt.category)
			if got != tt.want {
				t.Errorf("IsValidCategory(%q) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

// TestSortExpenses tests sorting expenses by amount and expense_date.
func TestSortExpenses(t *testing.T) {
	tests := []struct {
		name        string
		sortBy      string
		sortOrder   string
		wantFirstID int
	}{
		{
			name:        "sort by amount ascending",
			sortBy:      "amount",
			sortOrder:   "asc",
			wantFirstID: 2,
		},
		{
			name:        "sort by amount descending",
			sortBy:      "amount",
			sortOrder:   "desc",
			wantFirstID: 3,
		},
		{
			name:        "sort by date ascending",
			sortBy:      "expense_date",
			sortOrder:   "asc",
			wantFirstID: 3,
		},
		{
			name:        "sort by date descending",
			sortBy:      "expense_date",
			sortOrder:   "desc",
			wantFirstID: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fresh slice for every sub-test to avoid mutation between runs
			expenses := []models.Expense{
				{ID: 1, Amount: 350.50, ExpenseDate: "2025-06-10"},
				{ID: 2, Amount: 50.00, ExpenseDate: "2025-06-11"},
				{ID: 3, Amount: 800.00, ExpenseDate: "2025-06-09"},
			}

			models.SortExpenses(expenses, tt.sortBy, tt.sortOrder)

			if expenses[0].ID != tt.wantFirstID {
				t.Errorf("Expected first ID %d, got %d", tt.wantFirstID, expenses[0].ID)
			}
		})
	}
}

// TestGetNextExpenseID tests that the next ID is always max existing ID + 1.
func TestGetNextExpenseID(t *testing.T) {
	cleanup := setupTestExpensesCSV(t)
	defer cleanup()

	tests := []struct {
		name   string
		wantID int
	}{
		{
			name:   "returns next available ID",
			wantID: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := models.GetNextExpenseID()
			if id != tt.wantID {
				t.Errorf("Expected ID %d, got %d", tt.wantID, id)
			}
		})
	}
}
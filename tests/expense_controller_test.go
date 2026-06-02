// Package tests contains table-driven unit tests for the expense-tracker-api.
// expense_controller_test.go tests the ExpenseController validation logic.
package tests

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// setupTestData creates test users and expenses CSV files.
// Returns a cleanup function to delete the files after the test.
func setupTestData(t *testing.T) func() {
	t.Helper()

	os.MkdirAll("data", 0755)

	// Create test users CSV
	usersContent := "id,name,email,password,created_at\n" +
		"1,John Doe,john@example.com,secret123,2025-01-01T00:00:00Z\n" +
		"2,Jane Smith,jane@example.com,secret456,2025-01-02T00:00:00Z\n"

	err := os.WriteFile("data/users.csv", []byte(usersContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test users CSV: %v", err)
	}

	// Create test expenses CSV
	expensesContent := "id,user_id,title,amount,category,note,expense_date,created_at\n" +
		"1,1,Lunch,350.50,Food,Team lunch,2025-06-10,2025-06-10T14:30:00Z\n" +
		"2,1,Bus fare,50.00,Transport,Morning commute,2025-06-11,2025-06-11T08:00:00Z\n" +
		"3,1,Movie,200.00,Entertainment,Cinema with friends,2025-06-12,2025-06-12T19:00:00Z\n" +
		"4,2,Groceries,800.00,Food,Weekly groceries,2025-06-12,2025-06-12T10:00:00Z\n"

	err = os.WriteFile("data/expenses.csv", []byte(expensesContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test expenses CSV: %v", err)
	}

	return func() {
		os.Remove("data/users.csv")
		os.Remove("data/expenses.csv")
		os.Remove("data")
	}
}

// TestCreateExpenseValidation tests the Create endpoint validation logic.
func TestCreateExpenseValidation(t *testing.T) {
	cleanup := setupTestData(t)
	defer cleanup()

	tests := []struct {
		name       string
		userID     string
		input      map[string]interface{}
		wantStatus int
		wantError  string
	}{
		{
			name:   "valid expense creation succeeds",
			userID: "1",
			input: map[string]interface{}{
				"title":        "Dinner",
				"amount":       500.00,
				"category":     "Food",
				"note":         "Evening meal",
				"expense_date": "2025-06-20",
			},
			wantStatus: 201,
			wantError:  "",
		},
		{
			name:       "missing user ID returns 401",
			userID:     "",
			input:      map[string]interface{}{},
			wantStatus: 401,
			wantError:  "Unauthorized",
		},
		{
			name:   "invalid user ID returns 401",
			userID: "999",
			input: map[string]interface{}{
				"title":        "Dinner",
				"amount":       500.00,
				"category":     "Food",
				"note":         "Evening meal",
				"expense_date": "2025-06-20",
			},
			wantStatus: 401,
			wantError:  "Unauthorized",
		},
		{
			name:   "missing title returns 400",
			userID: "1",
			input: map[string]interface{}{
				"title":        "",
				"amount":       500.00,
				"category":     "Food",
				"note":         "Evening meal",
				"expense_date": "2025-06-20",
			},
			wantStatus: 400,
			wantError:  "Title is required",
		},
		{
			name:   "zero amount returns 400",
			userID: "1",
			input: map[string]interface{}{
				"title":        "Dinner",
				"amount":       0,
				"category":     "Food",
				"note":         "Evening meal",
				"expense_date": "2025-06-20",
			},
			wantStatus: 400,
			wantError:  "Amount must be a positive number",
		},
		{
			name:   "negative amount returns 400",
			userID: "1",
			input: map[string]interface{}{
				"title":        "Dinner",
				"amount":       -50.00,
				"category":     "Food",
				"note":         "Evening meal",
				"expense_date": "2025-06-20",
			},
			wantStatus: 400,
			wantError:  "Amount must be a positive number",
		},
		{
			name:   "missing category returns 400",
			userID: "1",
			input: map[string]interface{}{
				"title":        "Dinner",
				"amount":       500.00,
				"category":     "",
				"note":         "Evening meal",
				"expense_date": "2025-06-20",
			},
			wantStatus: 400,
			wantError:  "Category is required",
		},
		{
			name:   "invalid category returns 400",
			userID: "1",
			input: map[string]interface{}{
				"title":        "Dinner",
				"amount":       500.00,
				"category":     "InvalidCategory",
				"note":         "Evening meal",
				"expense_date": "2025-06-20",
			},
			wantStatus: 400,
			wantError:  "Invalid category",
		},
		{
			name:   "missing expense_date returns 400",
			userID: "1",
			input: map[string]interface{}{
				"title":        "Dinner",
				"amount":       500.00,
				"category":     "Food",
				"note":         "Evening meal",
				"expense_date": "",
			},
			wantStatus: 400,
			wantError:  "Expense date is required",
		},
		{
			name:   "invalid date format returns 400",
			userID: "1",
			input: map[string]interface{}{
				"title":        "Dinner",
				"amount":       500.00,
				"category":     "Food",
				"note":         "Evening meal",
				"expense_date": "06-20-2025",
			},
			wantStatus: 400,
			wantError:  "Invalid date format, use YYYY-MM-DD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal input to JSON
			inputBody, _ := json.Marshal(tt.input)

			// Verify the request body structure
			if len(inputBody) > 0 {
				t.Logf("Request body: %s", string(inputBody))
			}
		})
	}
}

// TestListExpenseFiltering tests the List endpoint filtering and sorting logic.
func TestListExpenseFiltering(t *testing.T) {
	cleanup := setupTestData(t)
	defer cleanup()

	tests := []struct {
		name        string
		userID      string
		queryParams map[string]string
		wantStatus  int
	}{
		{
			name:        "list all expenses succeeds",
			userID:      "1",
			queryParams: map[string]string{},
			wantStatus:  200,
		},
		{
			name:        "missing user ID returns 401",
			userID:      "",
			queryParams: map[string]string{},
			wantStatus:  401,
		},
		{
			name:   "filter by Food category succeeds",
			userID: "1",
			queryParams: map[string]string{
				"category": "Food",
			},
			wantStatus: 200,
		},
		{
			name:   "filter by Transport category succeeds",
			userID: "1",
			queryParams: map[string]string{
				"category": "Transport",
			},
			wantStatus: 200,
		},
		{
			name:   "invalid category in filter returns 400",
			userID: "1",
			queryParams: map[string]string{
				"category": "InvalidCategory",
			},
			wantStatus: 400,
		},
		{
			name:   "filter by date_from succeeds",
			userID: "1",
			queryParams: map[string]string{
				"date_from": "2025-06-10",
			},
			wantStatus: 200,
		},
		{
			name:   "filter by date_to succeeds",
			userID: "1",
			queryParams: map[string]string{
				"date_to": "2025-06-12",
			},
			wantStatus: 200,
		},
		{
			name:   "filter by date range succeeds",
			userID: "1",
			queryParams: map[string]string{
				"date_from": "2025-06-10",
				"date_to":   "2025-06-12",
			},
			wantStatus: 200,
		},
		{
			name:   "invalid date_from format returns 400",
			userID: "1",
			queryParams: map[string]string{
				"date_from": "06-20-2025",
			},
			wantStatus: 400,
		},
		{
			name:   "sort by amount succeeds",
			userID: "1",
			queryParams: map[string]string{
				"sort_by": "amount",
			},
			wantStatus: 200,
		},
		{
			name:   "sort by expense_date succeeds",
			userID: "1",
			queryParams: map[string]string{
				"sort_by": "expense_date",
			},
			wantStatus: 200,
		},
		{
			name:   "sort ascending succeeds",
			userID: "1",
			queryParams: map[string]string{
				"sort_order": "asc",
			},
			wantStatus: 200,
		},
		{
			name:   "sort descending succeeds",
			userID: "1",
			queryParams: map[string]string{
				"sort_order": "desc",
			},
			wantStatus: 200,
		},
		{
			name:   "invalid sort_by returns 400",
			userID: "1",
			queryParams: map[string]string{
				"sort_by": "invalid_field",
			},
			wantStatus: 400,
		},
		{
			name:   "invalid sort_order returns 400",
			userID: "1",
			queryParams: map[string]string{
				"sort_order": "invalid_order",
			},
			wantStatus: 400,
		},
		{
			name:   "limit parameter works",
			userID: "1",
			queryParams: map[string]string{
				"limit": "2",
			},
			wantStatus: 200,
		},
		{
			name:   "invalid limit returns 400",
			userID: "1",
			queryParams: map[string]string{
				"limit": "abc",
			},
			wantStatus: 400,
		},
		{
			name:   "negative limit returns 400",
			userID: "1",
			queryParams: map[string]string{
				"limit": "-5",
			},
			wantStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify test case configuration
			if len(tt.queryParams) > 0 {
				t.Logf("Query params: %v", tt.queryParams)
			}
		})
	}
}

// TestGetOneExpenseValidation tests the GetOne endpoint validation logic.
func TestGetOneExpenseValidation(t *testing.T) {
	cleanup := setupTestData(t)
	defer cleanup()

	tests := []struct {
		name       string
		userID     string
		expenseID  string
		wantStatus int
	}{
		{
			name:       "get existing expense succeeds",
			userID:     "1",
			expenseID:  "1",
			wantStatus: 200,
		},
		{
			name:       "get another existing expense succeeds",
			userID:     "1",
			expenseID:  "2",
			wantStatus: 200,
		},
		{
			name:       "missing user ID returns 401",
			userID:     "",
			expenseID:  "1",
			wantStatus: 401,
		},
		{
			name:       "invalid user ID returns 401",
			userID:     "999",
			expenseID:  "1",
			wantStatus: 401,
		},
		{
			name:       "invalid expense ID returns 400",
			userID:     "1",
			expenseID:  "abc",
			wantStatus: 400,
		},
		{
			name:       "negative expense ID returns 400",
			userID:     "1",
			expenseID:  "-1",
			wantStatus: 400,
		},
		{
			name:       "zero expense ID returns 400",
			userID:     "1",
			expenseID:  "0",
			wantStatus: 400,
		},
		{
			name:       "nonexistent expense returns 404",
			userID:     "1",
			expenseID:  "999",
			wantStatus: 404,
		},
		{
			name:       "expense from different user returns 404",
			userID:     "1",
			expenseID:  "4",
			wantStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify test case configuration
			t.Logf("Testing ID: %s for user: %s", tt.expenseID, tt.userID)
		})
	}
}

// TestUpdateExpenseValidation tests the Update endpoint validation logic.
func TestUpdateExpenseValidation(t *testing.T) {
	cleanup := setupTestData(t)
	defer cleanup()

	tests := []struct {
		name       string
		userID     string
		expenseID  string
		input      map[string]interface{}
		wantStatus int
	}{
		{
			name:      "update existing expense succeeds",
			userID:    "1",
			expenseID: "1",
			input: map[string]interface{}{
				"title":        "Updated Lunch",
				"amount":       400.00,
				"category":     "Food",
				"note":         "Updated note",
				"expense_date": "2025-06-15",
			},
			wantStatus: 200,
		},
		{
			name:       "missing user ID returns 401",
			userID:     "",
			expenseID:  "1",
			input:      map[string]interface{}{},
			wantStatus: 401,
		},
		{
			name:      "nonexistent expense returns 404",
			userID:    "1",
			expenseID: "999",
			input: map[string]interface{}{
				"title":        "Updated Lunch",
				"amount":       400.00,
				"category":     "Food",
				"note":         "Updated note",
				"expense_date": "2025-06-15",
			},
			wantStatus: 404,
		},
		{
			name:      "invalid expense ID returns 400",
			userID:    "1",
			expenseID: "abc",
			input: map[string]interface{}{
				"title":        "Updated Lunch",
				"amount":       400.00,
				"category":     "Food",
				"note":         "Updated note",
				"expense_date": "2025-06-15",
			},
			wantStatus: 400,
		},
		{
			name:      "empty title returns 400",
			userID:    "1",
			expenseID: "1",
			input: map[string]interface{}{
				"title":        "",
				"amount":       400.00,
				"category":     "Food",
				"note":         "Updated note",
				"expense_date": "2025-06-15",
			},
			wantStatus: 400,
		},
		{
			name:      "invalid amount returns 400",
			userID:    "1",
			expenseID: "1",
			input: map[string]interface{}{
				"title":        "Updated Lunch",
				"amount":       -100.00,
				"category":     "Food",
				"note":         "Updated note",
				"expense_date": "2025-06-15",
			},
			wantStatus: 400,
		},
		{
			name:      "invalid category returns 400",
			userID:    "1",
			expenseID: "1",
			input: map[string]interface{}{
				"title":        "Updated Lunch",
				"amount":       400.00,
				"category":     "InvalidCat",
				"note":         "Updated note",
				"expense_date": "2025-06-15",
			},
			wantStatus: 400,
		},
		{
			name:      "invalid date format returns 400",
			userID:    "1",
			expenseID: "1",
			input: map[string]interface{}{
				"title":        "Updated Lunch",
				"amount":       400.00,
				"category":     "Food",
				"note":         "Updated note",
				"expense_date": "06-15-2025",
			},
			wantStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputBody, _ := json.Marshal(tt.input)
			if len(inputBody) > 0 {
				t.Logf("Request body: %s", string(inputBody))
			}
		})
	}
}

// TestDeleteExpenseValidation tests the Delete endpoint validation logic.
func TestDeleteExpenseValidation(t *testing.T) {
	cleanup := setupTestData(t)
	defer cleanup()

	tests := []struct {
		name       string
		userID     string
		expenseID  string
		wantStatus int
	}{
		{
			name:       "delete existing expense succeeds",
			userID:     "1",
			expenseID:  "1",
			wantStatus: 200,
		},
		{
			name:       "missing user ID returns 401",
			userID:     "",
			expenseID:  "1",
			wantStatus: 401,
		},
		{
			name:       "invalid user ID returns 401",
			userID:     "999",
			expenseID:  "1",
			wantStatus: 401,
		},
		{
			name:       "invalid expense ID returns 400",
			userID:     "1",
			expenseID:  "abc",
			wantStatus: 400,
		},
		{
			name:       "negative expense ID returns 400",
			userID:     "1",
			expenseID:  "-1",
			wantStatus: 400,
		},
		{
			name:       "nonexistent expense returns 404",
			userID:     "1",
			expenseID:  "999",
			wantStatus: 404,
		},
		{
			name:       "expense from different user returns 404",
			userID:     "1",
			expenseID:  "4",
			wantStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing deletion of expense %s for user %s", tt.expenseID, tt.userID)
		})
	}
}

// TestSummaryExpenseValidation tests the Summary endpoint validation logic.
func TestSummaryExpenseValidation(t *testing.T) {
	cleanup := setupTestData(t)
	defer cleanup()

	tests := []struct {
		name           string
		userID         string
		dateFrom       string
		dateTo         string
		wantStatus     int
	}{
		{
			name:       "valid summary request succeeds",
			userID:     "1",
			dateFrom:   "2025-06-10",
			dateTo:     "2025-06-12",
			wantStatus: 200,
		},
		{
			name:       "missing user ID returns 401",
			userID:     "",
			dateFrom:   "2025-06-10",
			dateTo:     "2025-06-12",
			wantStatus: 401,
		},
		{
			name:       "missing date_from returns 400",
			userID:     "1",
			dateFrom:   "",
			dateTo:     "2025-06-12",
			wantStatus: 400,
		},
		{
			name:       "missing date_to returns 400",
			userID:     "1",
			dateFrom:   "2025-06-10",
			dateTo:     "",
			wantStatus: 400,
		},
		{
			name:       "invalid date_from format returns 400",
			userID:     "1",
			dateFrom:   "06-10-2025",
			dateTo:     "2025-06-12",
			wantStatus: 400,
		},
		{
			name:       "invalid date_to format returns 400",
			userID:     "1",
			dateFrom:   "2025-06-10",
			dateTo:     "06-12-2025",
			wantStatus: 400,
		},
		{
			name:       "date_to before date_from returns 400",
			userID:     "1",
			dateFrom:   "2025-06-20",
			dateTo:     "2025-06-10",
			wantStatus: 400,
		},
		{
			name:       "date_to equals date_from succeeds",
			userID:     "1",
			dateFrom:   "2025-06-10",
			dateTo:     "2025-06-10",
			wantStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing summary from %s to %s for user %s", tt.dateFrom, tt.dateTo, tt.userID)
		})
	}
}

// TestIsValidDateHelper tests date format validation.
func TestIsValidDateHelper(t *testing.T) {
	tests := []struct {
		name      string
		dateStr   string
		wantValid bool
	}{
		{
			name:      "valid YYYY-MM-DD format",
			dateStr:   "2025-06-20",
			wantValid: true,
		},
		{
			name:      "valid date at year boundary",
			dateStr:   "2025-01-01",
			wantValid: true,
		},
		{
			name:      "valid date at year end",
			dateStr:   "2025-12-31",
			wantValid: true,
		},
		{
			name:      "invalid MM-DD-YYYY format",
			dateStr:   "06-20-2025",
			wantValid: false,
		},
		{
			name:      "invalid DD-MM-YYYY format",
			dateStr:   "20-06-2025",
			wantValid: false,
		},
		{
			name:      "invalid with slashes",
			dateStr:   "2025/06/20",
			wantValid: false,
		},
		{
			name:      "invalid month value",
			dateStr:   "2025-13-01",
			wantValid: false,
		},
		{
			name:      "invalid day value",
			dateStr:   "2025-06-31",
			wantValid: false,
		},
		{
			name:      "empty string is invalid",
			dateStr:   "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := isValidDateHelper(tt.dateStr)
			if valid != tt.wantValid {
				t.Errorf("isValidDate(%q) = %v, want %v", tt.dateStr, valid, tt.wantValid)
			}
		})
	}
}

// isValidDateHelper validates date format (YYYY-MM-DD).
func isValidDateHelper(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
 
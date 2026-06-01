// Package models handles all data structures and CSV file operations.
// expense.go defines the Expense model and all expense-related CSV functions.
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

// Expense represents a single expense entry in the system.
// All expense data is stored and read from a CSV file.
type Expense struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	ExpenseDate string  `json:"expense_date"`
	CreatedAt   string  `json:"created_at"`
}

// AllowedCategories defines the valid expense categories.
var AllowedCategories = []string{
	"Food", "Transport", "Housing", "Entertainment",
	"Shopping", "Healthcare", "Education", "Utilities", "Other",
}

// IsValidCategory checks if the given category is in the allowed list.
func IsValidCategory(category string) bool {
	for _, c := range AllowedCategories {
		if c == category {
			return true
		}
	}
	return false
}

// getExpensesCSVPath reads the CSV file path from app.conf.
// Falls back to a default path if config is not set.
func getExpensesCSVPath() string {
	path, _ := beego.AppConfig.String("expenses_csv_path")
	if path == "" {
		return "data/expenses.csv"
	}
	return path
}

// GetAllExpenses reads all expenses from the CSV file and returns them as a slice.
// Returns an empty slice if the file does not exist yet.
func GetAllExpenses() ([]Expense, error) {
	filePath := getExpensesCSVPath()

	// If file does not exist yet, return empty list
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []Expense{}, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		logs.Error("Failed to open expenses CSV:", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		logs.Error("Failed to read expenses CSV:", err)
		return nil, err
	}

	var expenses []Expense

	// Skip header row at index 0
	for i, record := range records {
		if i == 0 {
			continue
		}

		// Parse all numeric fields safely
		id, err := strconv.Atoi(record[0])
		if err != nil {
			logs.Warn("Skipping expense row with invalid ID:", record)
			continue
		}

		userID, err := strconv.Atoi(record[1])
		if err != nil {
			logs.Warn("Skipping expense row with invalid UserID:", record)
			continue
		}

		amount, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			logs.Warn("Skipping expense row with invalid Amount:", record)
			continue
		}

		expenses = append(expenses, Expense{
			ID:          id,
			UserID:      userID,
			Title:       record[2],
			Amount:      amount,
			Category:    record[4],
			Note:        record[5],
			ExpenseDate: record[6],
			CreatedAt:   record[7],
		})
	}

	return expenses, nil
}

// GetExpensesByUserID returns all expenses belonging to a specific user.
func GetExpensesByUserID(userID int) ([]Expense, error) {
	all, err := GetAllExpenses()
	if err != nil {
		return nil, err
	}

	var userExpenses []Expense
	for _, e := range all {
		if e.UserID == userID {
			userExpenses = append(userExpenses, e)
		}
	}

	return userExpenses, nil
}

// GetExpenseByID finds a single expense by ID that belongs to the given user.
// Returns an error if not found or if the expense belongs to another user.
func GetExpenseByID(id int, userID int) (*Expense, error) {
	all, err := GetAllExpenses()
	if err != nil {
		return nil, err
	}

	for _, e := range all {
		if e.ID == id && e.UserID == userID {
			return &e, nil
		}
	}

	return nil, errors.New("expense not found")
}

// GetNextExpenseID determines the next available expense ID
// by finding the maximum existing ID and adding 1.
func GetNextExpenseID() int {
	all, err := GetAllExpenses()
	if err != nil || len(all) == 0 {
		return 1
	}

	maxID := 0
	for _, e := range all {
		if e.ID > maxID {
			maxID = e.ID
		}
	}

	return maxID + 1
}

// CreateExpense appends a new expense record to the CSV file.
// Creates the file with a header row if it does not exist yet.
func CreateExpense(expense *Expense) error {
	filePath := getExpensesCSVPath()

	// Ensure the data directory exists before writing
	if err := os.MkdirAll("data", 0755); err != nil {
		logs.Error("Failed to create data directory:", err)
		return err
	}

	// Check if file already exists to avoid writing header twice
	fileExists := true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fileExists = false
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logs.Error("Failed to open expenses CSV for writing:", err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header only when creating the file for the first time
	if !fileExists {
		header := []string{"id", "user_id", "title", "amount", "category", "note", "expense_date", "created_at"}
		if err := writer.Write(header); err != nil {
			logs.Error("Failed to write expenses CSV header:", err)
			return err
		}
	}

	// Set creation timestamp if not already provided
	if expense.CreatedAt == "" {
		expense.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	record := []string{
		strconv.Itoa(expense.ID),
		strconv.Itoa(expense.UserID),
		expense.Title,
		strconv.FormatFloat(expense.Amount, 'f', 2, 64),
		expense.Category,
		expense.Note,
		expense.ExpenseDate,
		expense.CreatedAt,
	}

	if err := writer.Write(record); err != nil {
		logs.Error("Failed to write expense record to CSV:", err)
		return err
	}

	logs.Info("New expense created with ID:", expense.ID)
	return nil
}

// UpdateExpense rewrites the entire CSV file with the updated expense record.
// Returns an error if the expense is not found.
func UpdateExpense(updated *Expense) error {
	all, err := GetAllExpenses()
	if err != nil {
		return err
	}

	// Find and replace the matching expense
	found := false
	for i, e := range all {
		if e.ID == updated.ID && e.UserID == updated.UserID {
			all[i] = *updated
			found = true
			break
		}
	}

	if !found {
		return errors.New("expense not found")
	}

	// Rewrite entire CSV with the updated data
	if err := writeAllExpenses(all); err != nil {
		return err
	}

	logs.Info("Expense updated with ID:", updated.ID)
	return nil
}

// DeleteExpense rewrites the CSV file excluding the deleted expense row.
// Returns an error if the expense is not found.
func DeleteExpense(id int, userID int) error {
	all, err := GetAllExpenses()
	if err != nil {
		return err
	}

	// Build new slice without the deleted expense
	var remaining []Expense
	found := false
	for _, e := range all {
		if e.ID == id && e.UserID == userID {
			found = true
			continue
		}
		remaining = append(remaining, e)
	}

	if !found {
		return errors.New("expense not found")
	}

	// Rewrite entire CSV without the deleted row
	if err := writeAllExpenses(remaining); err != nil {
		return err
	}

	logs.Info("Expense deleted with ID:", id)
	return nil
}

// writeAllExpenses rewrites the entire expenses CSV file from a slice.
// This is used internally by UpdateExpense and DeleteExpense.
func writeAllExpenses(expenses []Expense) error {
	filePath := getExpensesCSVPath()

	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logs.Error("Failed to open expenses CSV for rewriting:", err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Always write header first
	header := []string{"id", "user_id", "title", "amount", "category", "note", "expense_date", "created_at"}
	if err := writer.Write(header); err != nil {
		logs.Error("Failed to write expenses CSV header:", err)
		return err
	}

	// Write all expense rows
	for _, e := range expenses {
		record := []string{
			strconv.Itoa(e.ID),
			strconv.Itoa(e.UserID),
			e.Title,
			strconv.FormatFloat(e.Amount, 'f', 2, 64),
			e.Category,
			e.Note,
			e.ExpenseDate,
			e.CreatedAt,
		}
		if err := writer.Write(record); err != nil {
			logs.Error("Failed to write expense row during rewrite:", err)
			return err
		}
	}

	return nil
}

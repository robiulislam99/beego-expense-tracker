// Package controllers handles all incoming HTTP requests and sends responses.
// expense.go handles all expense CRUD endpoints.
package controllers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"expense-tracker-api/models"

	"github.com/beego/beego/v2/core/logs"
)

// ExpenseController handles all expense-related endpoints
// including create, read, update, and delete operations.
type ExpenseController struct {
	BaseController
}

// expenseInput defines the expected JSON body for create and update endpoints.
type expenseInput struct {
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	ExpenseDate string  `json:"expense_date"`
}

// getCurrentUserID reads and validates the X-User-ID header.
// Returns the user ID and true if valid, or 0 and false if missing or invalid.
func (c *ExpenseController) getCurrentUserID() (int, bool) {
	header := c.Ctx.Input.Header("X-User-ID")
	if header == "" {
		c.SendError(401, "Unauthorized")
		return 0, false
	}

	userID, err := strconv.Atoi(header)
	if err != nil || userID <= 0 {
		c.SendError(401, "Unauthorized")
		return 0, false
	}

	// Verify user exists in CSV
	_, err = models.GetUserByID(userID)
	if err != nil {
		logs.Warn("ExpenseController: user not found for ID:", userID)
		c.SendError(401, "Unauthorized")
		return 0, false
	}

	return userID, true
}

// Create handles POST /api/v1/expenses
// Creates a new expense for the authenticated user.
func (c *ExpenseController) Create() {
	logs.Info("Create expense endpoint called")

	userID, ok := c.getCurrentUserID()
	if !ok {
		return
	}

	var input expenseInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Create expense: invalid JSON body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	// Trim whitespace
	input.Title = strings.TrimSpace(input.Title)
	input.Category = strings.TrimSpace(input.Category)
	input.Note = strings.TrimSpace(input.Note)
	input.ExpenseDate = strings.TrimSpace(input.ExpenseDate)

	// Validate title
	if input.Title == "" {
		c.SendError(400, "Title is required")
		return
	}

	// Validate amount
	if input.Amount <= 0 {
		c.SendError(400, "Amount must be a positive number")
		return
	}

	// Validate category
	if input.Category == "" {
		c.SendError(400, "Category is required")
		return
	}
	if !models.IsValidCategory(input.Category) {
		c.SendError(400, "Invalid category")
		return
	}

	// Validate expense_date format YYYY-MM-DD
	if input.ExpenseDate == "" {
		c.SendError(400, "Expense date is required")
		return
	}
	if !isValidDate(input.ExpenseDate) {
		c.SendError(400, "Invalid date format, use YYYY-MM-DD")
		return
	}

	newExpense := &models.Expense{
		ID:          models.GetNextExpenseID(),
		UserID:      userID,
		Title:       input.Title,
		Amount:      input.Amount,
		Category:    input.Category,
		Note:        input.Note,
		ExpenseDate: input.ExpenseDate,
	}

	if err := models.CreateExpense(newExpense); err != nil {
		logs.Error("Create expense: failed to save:", err)
		c.SendError(500, "Failed to create expense")
		return
	}

	logs.Info("Expense created with ID:", newExpense.ID)
	c.SendCreated("Expense created successfully", newExpense)
}

// List handles GET /api/v1/expenses
// Returns all expenses for the authenticated user with optional pagination.
func (c *ExpenseController) List() {
	logs.Info("List expenses endpoint called")

	userID, ok := c.getCurrentUserID()
	if !ok {
		return
	}

	expenses, err := models.GetExpensesByUserID(userID)
	if err != nil {
		logs.Error("List expenses: failed to read CSV:", err)
		c.SendError(500, "Failed to retrieve expenses")
		return
	}

	// Apply pagination via ?limit query parameter
	limitStr := c.GetString("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.SendError(400, "Invalid limit parameter")
			return
		}
		if limit < len(expenses) {
			expenses = expenses[:limit]
		}
	}

	// Return empty array instead of null when no expenses found
	if expenses == nil {
		expenses = []models.Expense{}
	}

	c.SendSuccess("Expenses retrieved", expenses)
}

// GetOne handles GET /api/v1/expenses/:id
// Returns a single expense by ID for the authenticated user.
func (c *ExpenseController) GetOne() {
	logs.Info("Get one expense endpoint called")

	userID, ok := c.getCurrentUserID()
	if !ok {
		return
	}

	// Parse :id from URL
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.SendError(400, "Invalid expense ID")
		return
	}

	expense, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Warn("Get one expense: not found, ID:", id)
		c.SendError(404, "Expense not found")
		return
	}

	c.SendSuccess("Expense retrieved", expense)
}

// Update handles PUT /api/v1/expenses/:id
// Updates an existing expense for the authenticated user.
func (c *ExpenseController) Update() {
	logs.Info("Update expense endpoint called")

	userID, ok := c.getCurrentUserID()
	if !ok {
		return
	}

	// Parse :id from URL
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.SendError(400, "Invalid expense ID")
		return
	}

	// Check expense exists and belongs to user
	existing, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Warn("Update expense: not found, ID:", id)
		c.SendError(404, "Expense not found")
		return
	}

	var input expenseInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Update expense: invalid JSON body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	// Trim whitespace
	input.Title = strings.TrimSpace(input.Title)
	input.Category = strings.TrimSpace(input.Category)
	input.Note = strings.TrimSpace(input.Note)
	input.ExpenseDate = strings.TrimSpace(input.ExpenseDate)

	// Validate fields
	if input.Title == "" {
		c.SendError(400, "Title is required")
		return
	}
	if input.Amount <= 0 {
		c.SendError(400, "Amount must be a positive number")
		return
	}
	if input.Category == "" {
		c.SendError(400, "Category is required")
		return
	}
	if !models.IsValidCategory(input.Category) {
		c.SendError(400, "Invalid category")
		return
	}
	if input.ExpenseDate == "" {
		c.SendError(400, "Expense date is required")
		return
	}
	if !isValidDate(input.ExpenseDate) {
		c.SendError(400, "Invalid date format, use YYYY-MM-DD")
		return
	}

	// Apply updates while keeping original ID, UserID, and CreatedAt
	existing.Title = input.Title
	existing.Amount = input.Amount
	existing.Category = input.Category
	existing.Note = input.Note
	existing.ExpenseDate = input.ExpenseDate

	if err := models.UpdateExpense(existing); err != nil {
		logs.Error("Update expense: failed to save:", err)
		c.SendError(500, "Failed to update expense")
		return
	}

	logs.Info("Expense updated with ID:", id)
	c.SendSuccess("Expense updated successfully", existing)
}

// Delete handles DELETE /api/v1/expenses/:id
// Deletes an expense by ID for the authenticated user.
func (c *ExpenseController) Delete() {
	logs.Info("Delete expense endpoint called")

	userID, ok := c.getCurrentUserID()
	if !ok {
		return
	}

	// Parse :id from URL
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.SendError(400, "Invalid expense ID")
		return
	}

	// Check expense exists before deleting
	_, err = models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Warn("Delete expense: not found, ID:", id)
		c.SendError(404, "Expense not found")
		return
	}

	if err := models.DeleteExpense(id, userID); err != nil {
		logs.Error("Delete expense: failed:", err)
		c.SendError(500, "Failed to delete expense")
		return
	}

	logs.Info("Expense deleted with ID:", id)
	c.SendSuccess("Expense deleted successfully", nil)
}

// isValidDate checks if the string matches the YYYY-MM-DD format.
func isValidDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

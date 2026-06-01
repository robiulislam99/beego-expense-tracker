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
// Returns expenses for the authenticated user with optional filtering, sorting, and pagination.
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

	// --- Step 1: Filter by category ---
	category := strings.TrimSpace(c.GetString("category"))
	if category != "" {
		if !models.IsValidCategory(category) {
			c.SendError(400, "Invalid category")
			return
		}
		var filtered []models.Expense
		for _, e := range expenses {
			if e.Category == category {
				filtered = append(filtered, e)
			}
		}
		expenses = filteredgo test -v -coverpkg=./... ./tests/...
	}

	// --- Step 2: Filter by date range ---
	dateFrom := strings.TrimSpace(c.GetString("date_from"))
	dateTo := strings.TrimSpace(c.GetString("date_to"))

	if dateFrom != "" {
		if !isValidDate(dateFrom) {
			c.SendError(400, "Invalid date_from format, use YYYY-MM-DD")
			return
		}
		var filtered []models.Expense
		for _, e := range expenses {
			if e.ExpenseDate >= dateFrom {
				filtered = append(filtered, e)
			}
		}
		expenses = filtered
	}

	if dateTo != "" {
		if !isValidDate(dateTo) {
			c.SendError(400, "Invalid date_to format, use YYYY-MM-DD")
			return
		}
		var filtered []models.Expense
		for _, e := range expenses {
			if e.ExpenseDate <= dateTo {
				filtered = append(filtered, e)
			}
		}
		expenses = filtered
	}

	// --- Step 3: Sort ---
	sortBy := strings.TrimSpace(c.GetString("sort_by"))
	sortOrder := strings.TrimSpace(c.GetString("sort_order"))

	// Default sort order is descending
	if sortOrder == "" {
		sortOrder = "desc"
	}

	if sortBy != "" && sortBy != "amount" && sortBy != "expense_date" {
		c.SendError(400, "Invalid sort_by, use amount or expense_date")
		return
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		c.SendError(400, "Invalid sort_order, use asc or desc")
		return
	}

	if sortBy != "" {
		models.SortExpenses(expenses, sortBy, sortOrder)
	}

	// --- Step 4: Pagination via ?limit ---
	limitStr := strings.TrimSpace(c.GetString("limit"))
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

// Summary handles GET /api/v1/expenses/summary
// Returns total spending grouped by category for a given date range.
func (c *ExpenseController) Summary() {
	logs.Info("Summary endpoint called")

	userID, ok := c.getCurrentUserID()
	if !ok {
		return
	}

	// Both date_from and date_to are required
	dateFrom := strings.TrimSpace(c.GetString("date_from"))
	dateTo := strings.TrimSpace(c.GetString("date_to"))

	if dateFrom == "" {
		c.SendError(400, "date_from is required")
		return
	}
	if dateTo == "" {
		c.SendError(400, "date_to is required")
		return
	}
	if !isValidDate(dateFrom) {
		c.SendError(400, "Invalid date_from format, use YYYY-MM-DD")
		return
	}
	if !isValidDate(dateTo) {
		c.SendError(400, "Invalid date_to format, use YYYY-MM-DD")
		return
	}
	if dateTo < dateFrom {
		c.SendError(400, "date_to must be after date_from")
		return
	}

	expenses, err := models.GetExpensesByUserID(userID)
	if err != nil {
		logs.Error("Summary: failed to read CSV:", err)
		c.SendError(500, "Failed to retrieve expenses")
		return
	}

	// Filter by date range
	var filtered []models.Expense
	for _, e := range expenses {
		if e.ExpenseDate >= dateFrom && e.ExpenseDate <= dateTo {
			filtered = append(filtered, e)
		}
	}

	// Group totals and counts by category using a map
	categoryTotals := make(map[string]float64)
	categoryCounts := make(map[string]int)
	totalAmount := 0.0

	for _, e := range filtered {
		categoryTotals[e.Category] += e.Amount
		categoryCounts[e.Category]++
		totalAmount += e.Amount
	}

	// Build the by_category slice
	type categoryEntry struct {
		Category string  `json:"category"`
		Total    float64 `json:"total"`
		Count    int     `json:"count"`
	}

	var byCategory []categoryEntry
	for _, cat := range models.AllowedCategories {
		if total, exists := categoryTotals[cat]; exists {
			byCategory = append(byCategory, categoryEntry{
				Category: cat,
				Total:    total,
				Count:    categoryCounts[cat],
			})
		}
	}

	// Build summary response
	summary := map[string]interface{}{
		"date_from":    dateFrom,
		"date_to":      dateTo,
		"total_amount": totalAmount,
		"total_count":  len(filtered),
		"by_category":  byCategory,
	}

	logs.Info("Summary generated for user ID:", userID)
	c.SendSuccess("Summary generated", summary)
}

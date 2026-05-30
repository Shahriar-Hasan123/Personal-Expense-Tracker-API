package controllers

import (
	"encoding/json"
	"expense-tracker-api/models"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

// ExpenseController handles all expense-related endpoints.
type ExpenseController struct {
	BaseController
}

// expenseInput defines the expected JSON body for create and update.
type expenseInput struct {
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	ExpenseDate string  `json:"expense_date"`
}

// expenseResponse defines the shape of an expense in API responses.
type expenseResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	ExpenseDate string  `json:"expense_date"`
	CreatedAt   string  `json:"created_at"`
}

// toResponse converts an Expense model to the API response shape.
func toResponse(e *models.Expense) expenseResponse {
	return expenseResponse{
		ID:          e.ID,
		Title:       e.Title,
		Amount:      e.Amount,
		Category:    e.Category,
		Note:        e.Note,
		ExpenseDate: e.ExpenseDate,
		CreatedAt:   e.CreatedAt,
	}
}

// authenticate extracts and validates X-User-ID from the request header.
// Returns userID on success, or writes a 401 response and returns -1.
func (c *ExpenseController) authenticate() int {
	header := c.Ctx.Input.Header("X-User-ID")
	if header == "" {
		c.respondUnauthorized("Unauthorized")
		return -1
	}

	userID, err := strconv.Atoi(header)
	if err != nil || userID <= 0 {
		c.respondUnauthorized("Unauthorized")
		return -1
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		logs.Error("authenticate: GetUserByID error: %v", err)
		c.respondInternalError("Internal server error")
		return -1
	}
	if user == nil {
		c.respondUnauthorized("Unauthorized")
		return -1
	}

	return userID
}

// validateExpenseInput checks all required fields and business rules.
func validateExpenseInput(input *expenseInput) string {
	input.Title = strings.TrimSpace(input.Title)
	input.Category = strings.TrimSpace(input.Category)
	input.ExpenseDate = strings.TrimSpace(input.ExpenseDate)

	if input.Title == "" {
		return "Title is required"
	}
	if input.Amount <= 0 {
		return "Amount must be a positive number"
	}
	if input.Category == "" {
		return "Category is required"
	}
	if !models.IsValidCategory(input.Category) {
		return "Invalid category"
	}
	if input.ExpenseDate == "" {
		return "Expense date is required"
	}
	if _, err := time.Parse("2006-01-02", input.ExpenseDate); err != nil {
		return "Expense date must be in YYYY-MM-DD format"
	}
	return ""
}

// Create handles POST /api/v1/expenses
func (c *ExpenseController) Create() {
	userID := c.authenticate()
	if userID == -1 {
		return
	}

	var input expenseInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.respondBadRequest("Invalid request body")
		return
	}

	if errMsg := validateExpenseInput(&input); errMsg != "" {
		c.respondBadRequest(errMsg)
		return
	}

	expense := &models.Expense{
		ID:          models.GetNextExpenseID(),
		UserID:      userID,
		Title:       input.Title,
		Amount:      input.Amount,
		Category:    input.Category,
		Note:        input.Note,
		ExpenseDate: input.ExpenseDate,
	}

	if err := models.CreateExpense(expense); err != nil {
		logs.Error("Create expense error: %v", err)
		c.respondInternalError("Failed to create expense")
		return
	}

	logs.Info("Expense created: ID %d UserID %d", expense.ID, userID)
	c.respondCreated("Expense created successfully", toResponse(expense))
}

// List handles GET /api/v1/expenses
func (c *ExpenseController) List() {
	userID := c.authenticate()
	if userID == -1 {
		return
	}

	expenses, err := models.GetExpensesByUserID(userID)
	if err != nil {
		logs.Error("List expenses error: %v", err)
		c.respondInternalError("Failed to retrieve expenses")
		return
	}

	// Apply limit query parameter
	limitStr := c.GetString("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.respondBadRequest("limit must be a positive integer")
			return
		}
		if limit < len(expenses) {
			expenses = expenses[:limit]
		}
	}

	// Build response slice — never return null for empty list
	result := make([]expenseResponse, 0, len(expenses))
	for _, e := range expenses {
		e := e
		result = append(result, toResponse(&e))
	}

	c.respondSuccess("Expenses retrieved", result)
}

// GetOne handles GET /api/v1/expenses/:id
func (c *ExpenseController) GetOne() {
	userID := c.authenticate()
	if userID == -1 {
		return
	}

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil || id <= 0 {
		c.respondBadRequest("Invalid expense ID")
		return
	}

	expense, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Error("GetOne expense error: %v", err)
		c.respondInternalError("Failed to retrieve expense")
		return
	}
	if expense == nil {
		c.respondJSON(404, false, "Expense not found", nil)
		return
	}

	c.respondSuccess("Expense retrieved", toResponse(expense))
}

// Update handles PUT /api/v1/expenses/:id
func (c *ExpenseController) Update() {
	userID := c.authenticate()
	if userID == -1 {
		return
	}

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil || id <= 0 {
		c.respondBadRequest("Invalid expense ID")
		return
	}

	// Confirm ownership before updating
	existing, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Error("Update: GetExpenseByID error: %v", err)
		c.respondInternalError("Internal server error")
		return
	}
	if existing == nil {
		c.respondJSON(404, false, "Expense not found", nil)
		return
	}

	var input expenseInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.respondBadRequest("Invalid request body")
		return
	}

	if errMsg := validateExpenseInput(&input); errMsg != "" {
		c.respondBadRequest(errMsg)
		return
	}

	updated := &models.Expense{
		ID:          id,
		UserID:      userID,
		Title:       input.Title,
		Amount:      input.Amount,
		Category:    input.Category,
		Note:        input.Note,
		ExpenseDate: input.ExpenseDate,
	}

	if err := models.UpdateExpense(updated); err != nil {
		logs.Error("Update expense error: %v", err)
		c.respondInternalError("Failed to update expense")
		return
	}

	logs.Info("Expense updated: ID %d UserID %d", id, userID)
	c.respondSuccess("Expense updated successfully", toResponse(updated))
}

// Delete handles DELETE /api/v1/expenses/:id
func (c *ExpenseController) Delete() {
	userID := c.authenticate()
	if userID == -1 {
		return
	}

	id, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil || id <= 0 {
		c.respondBadRequest("Invalid expense ID")
		return
	}

	// Confirm ownership before deleting
	existing, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Error("Delete: GetExpenseByID error: %v", err)
		c.respondInternalError("Internal server error")
		return
	}
	if existing == nil {
		c.respondJSON(404, false, "Expense not found", nil)
		return
	}

	if err := models.DeleteExpense(id, userID); err != nil {
		logs.Error("Delete expense error: %v", err)
		c.respondInternalError("Failed to delete expense")
		return
	}

	logs.Info("Expense deleted: ID %d UserID %d", id, userID)
	c.respondSuccess("Expense deleted successfully", nil)
}

// Summary handles GET /api/v1/expenses/summary — implemented in Phase 3.
func (c *ExpenseController) Summary() {
	c.respondJSON(501, false, "Not implemented yet", nil)
}

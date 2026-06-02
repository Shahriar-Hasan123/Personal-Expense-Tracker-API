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

// parseListQueryParams reads, validates, and returns FilterOptions from the request.
// Returns an error message string if any param is invalid.
func (c *ExpenseController) parseListQueryParams() (models.FilterOptions, string) {
	opts := models.FilterOptions{}

	// category
	category := c.GetString("category")
	if category != "" && !models.IsValidCategory(category) {
		return opts, "Invalid category"
	}
	opts.Category = category

	// date_from
	dateFrom := c.GetString("date_from")
	if dateFrom != "" {
		if _, err := time.Parse("2006-01-02", dateFrom); err != nil {
			return opts, "date_from must be in YYYY-MM-DD format"
		}
	}
	opts.DateFrom = dateFrom

	// date_to
	dateTo := c.GetString("date_to")
	if dateTo != "" {
		if _, err := time.Parse("2006-01-02", dateTo); err != nil {
			return opts, "date_to must be in YYYY-MM-DD format"
		}
	}
	opts.DateTo = dateTo

	// date range logical check
	if dateFrom != "" && dateTo != "" && dateFrom > dateTo {
		return opts, "date_from cannot be after date_to"
	}

	// sort_by
	sortBy := c.GetString("sort_by")
	if sortBy != "" && sortBy != "amount" && sortBy != "expense_date" {
		return opts, "sort_by must be 'amount' or 'expense_date'"
	}
	opts.SortBy = sortBy

	// sort_order
	sortOrder := c.GetString("sort_order")
	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		return opts, "sort_order must be 'asc' or 'desc'"
	}
	opts.SortOrder = sortOrder

	// limit
	limitStr := c.GetString("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return opts, "limit must be a positive integer"
		}
		opts.Limit = limit
	}

	return opts, ""
}

// Create handles POST /api/v1/expenses
// @Title Create Expense
// @Summary Create a new expense
// @Description Creates a new expense for the authenticated user
// @Tags expenses
// @Accept json
// @Produce json
// @Param X-User-ID header int true "User ID"
// @Param body body expenseInput true "Expense details"
// @Success 201 {object} map[string]interface{} "Expense created successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /expenses [post]
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
// @Title List Expenses
// @Summary List all expenses for the user
// @Description Returns filtered, sorted, and paginated expenses
// @Tags expenses
// @Produce json
// @Param X-User-ID header int true "User ID"
// @Param category query string false "Filter by category"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Param sort_by query string false "Sort field: amount or expense_date"
// @Param sort_order query string false "Sort order: asc or desc"
// @Param limit query int false "Max number of results"
// @Success 200 {object} map[string]interface{} "Expenses retrieved"
// @Failure 400 {object} map[string]interface{} "Invalid query param"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /expenses [get]
func (c *ExpenseController) List() {
	userID := c.authenticate()
	if userID == -1 {
		return
	}

	opts, errMsg := c.parseListQueryParams()
	if errMsg != "" {
		c.respondBadRequest(errMsg)
		return
	}

	expenses, err := models.GetExpensesByUserID(userID)
	if err != nil {
		logs.Error("List expenses error: %v", err)
		c.respondInternalError("Failed to retrieve expenses")
		return
	}

	// Apply filters, sorting, and limit via model layer
	filtered := models.ApplyFilters(expenses, opts)

	// Build response slice — never return null for empty list
	result := make([]expenseResponse, 0, len(filtered))
	for i := range filtered {
		result = append(result, toResponse(&filtered[i]))
	}

	c.respondSuccess("Expenses retrieved", result)
}

// GetOne handles GET /api/v1/expenses/:id
// @Title Get Expense
// @Summary Get a single expense by ID
// @Description Returns one expense belonging to the authenticated user
// @Tags expenses
// @Produce json
// @Param X-User-ID header int true "User ID"
// @Param id path int true "Expense ID"
// @Success 200 {object} map[string]interface{} "Expense retrieved"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Expense not found"
// @Router /expenses/{id} [get]
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
// @Title Update Expense
// @Summary Update an existing expense
// @Description Updates an expense belonging to the authenticated user
// @Tags expenses
// @Accept json
// @Produce json
// @Param X-User-ID header int true "User ID"
// @Param id path int true "Expense ID"
// @Param body body expenseInput true "Updated expense details"
// @Success 200 {object} map[string]interface{} "Expense updated successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Expense not found"
// @Router /expenses/{id} [put]
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
// @Title Delete Expense
// @Summary Delete an expense
// @Description Deletes an expense belonging to the authenticated user
// @Tags expenses
// @Produce json
// @Param X-User-ID header int true "User ID"
// @Param id path int true "Expense ID"
// @Success 200 {object} map[string]interface{} "Expense deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Expense not found"
// @Router /expenses/{id} [delete]
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

// Summary handles GET /api/v1/expenses/summary
// @Title Expense Summary
// @Summary Get spending summary for a date range
// @Description Returns total spending and per-category breakdown
// @Tags expenses
// @Produce json
// @Param X-User-ID header int true "User ID"
// @Param date_from query string true "Start date (YYYY-MM-DD)"
// @Param date_to query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Summary generated"
// @Failure 400 {object} map[string]interface{} "Missing or invalid dates"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /expenses/summary [get]
func (c *ExpenseController) Summary() {
	userID := c.authenticate()
	if userID == -1 {
		return
	}

	// Both date params are required for summary
	dateFrom := strings.TrimSpace(c.GetString("date_from"))
	dateTo := strings.TrimSpace(c.GetString("date_to"))

	if dateFrom == "" {
		c.respondBadRequest("date_from is required")
		return
	}
	if dateTo == "" {
		c.respondBadRequest("date_to is required")
		return
	}
	if _, err := time.Parse("2006-01-02", dateFrom); err != nil {
		c.respondBadRequest("date_from must be in YYYY-MM-DD format")
		return
	}
	if _, err := time.Parse("2006-01-02", dateTo); err != nil {
		c.respondBadRequest("date_to must be in YYYY-MM-DD format")
		return
	}
	if dateFrom > dateTo {
		c.respondBadRequest("date_from cannot be after date_to")
		return
	}

	// Load all user expenses then filter to date range
	expenses, err := models.GetExpensesByUserID(userID)
	if err != nil {
		logs.Error("Summary: GetExpensesByUserID error: %v", err)
		c.respondInternalError("Failed to retrieve expenses")
		return
	}

	opts := models.FilterOptions{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}
	filtered := models.ApplyFilters(expenses, opts)

	summary := models.BuildSummary(filtered, dateFrom, dateTo)

	logs.Info("Summary generated for UserID %d range: %s to %s", userID, dateFrom, dateTo)
	c.respondSuccess("Summary generated", summary)
}

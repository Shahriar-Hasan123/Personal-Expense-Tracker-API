package controllers

import (
	"bytes"
	"encoding/json"
	"expense-tracker-api/models"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
	beegoContext "github.com/beego/beego/v2/server/web/context"
)

// setupExpenseCtrlTestEnv wires temp CSVs and returns cleanup.
func setupExpenseCtrlTestEnv(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	beego.AppConfig.Set("userscsv", filepath.Join(dir, "users.csv"))
	beego.AppConfig.Set("expensescsv", filepath.Join(dir, "expenses.csv"))
	beego.BConfig.WebConfig.AutoRender = false
	return func() { os.RemoveAll(dir) }
}

// execExpense wires an ExpenseController and calls the named method.
func execExpense(
	t *testing.T,
	method, url, body string,
	headers map[string]string,
	params map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rr := httptest.NewRecorder()
	ctx := beegoContext.NewContext()
	ctx.Reset(rr, req)
	ctx.Input.RequestBody = []byte(body)
	for k, v := range params {
		ctx.Input.SetParam(k, v)
	}

	c := &ExpenseController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	switch method {
	case "Create":
		c.Create()
	case "List":
		c.List()
	case "GetOne":
		c.GetOne()
	case "Update":
		c.Update()
	case "Delete":
		c.Delete()
	case "Summary":
		c.Summary()
	}
	return rr
}

// userHeader returns an X-User-ID header map.
func userHeader(id int) map[string]string {
	return map[string]string{"X-User-ID": fmt.Sprintf("%d", id)}
}

// seedTestUser creates a user directly via model for controller tests.
func seedTestUser(t *testing.T, id int, email string) *models.User {
	t.Helper()
	u := &models.User{ID: id, Name: "Test", Email: email, Password: "pass123"}
	if err := models.CreateUser(u); err != nil {
		t.Fatalf("seedTestUser: %v", err)
	}
	return u
}

// seedTestExpense creates an expense directly via model for controller tests.
func seedTestExpense(t *testing.T, id, userID int, title, category, date string, amount float64) *models.Expense {
	t.Helper()
	e := &models.Expense{
		ID: id, UserID: userID, Title: title,
		Amount: amount, Category: category, ExpenseDate: date,
	}
	if err := models.CreateExpense(e); err != nil {
		t.Fatalf("seedTestExpense: %v", err)
	}
	return e
}

// msg extracts the message field from a response body.
func msg(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("msg: decode error: %v — body: %s", err, rr.Body.String())
	}
	return result["message"].(string)
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreateExpense(t *testing.T) {
	defer setupExpenseCtrlTestEnv(t)()
	user := seedTestUser(t, 1, "alice@example.com")

	tests := []struct {
		name       string
		body       string
		headers    map[string]string
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "valid expense",
			body:       `{"title":"Lunch","amount":350.50,"category":"Food","expense_date":"2025-06-10"}`,
			headers:    userHeader(user.ID),
			wantStatus: 201,
			wantMsg:    "Expense created successfully",
		},
		{
			name:       "missing title",
			body:       `{"amount":350.50,"category":"Food","expense_date":"2025-06-10"}`,
			headers:    userHeader(user.ID),
			wantStatus: 400,
			wantMsg:    "Title is required",
		},
		{
			name:       "zero amount",
			body:       `{"title":"Lunch","amount":0,"category":"Food","expense_date":"2025-06-10"}`,
			headers:    userHeader(user.ID),
			wantStatus: 400,
			wantMsg:    "Amount must be a positive number",
		},
		{
			name:       "negative amount",
			body:       `{"title":"Lunch","amount":-10,"category":"Food","expense_date":"2025-06-10"}`,
			headers:    userHeader(user.ID),
			wantStatus: 400,
			wantMsg:    "Amount must be a positive number",
		},
		{
			name:       "invalid category",
			body:       `{"title":"Lunch","amount":350.50,"category":"InvalidCat","expense_date":"2025-06-10"}`,
			headers:    userHeader(user.ID),
			wantStatus: 400,
			wantMsg:    "Invalid category",
		},
		{
			name:       "missing expense_date",
			body:       `{"title":"Lunch","amount":350.50,"category":"Food"}`,
			headers:    userHeader(user.ID),
			wantStatus: 400,
			wantMsg:    "Expense date is required",
		},
		{
			name:       "invalid date format",
			body:       `{"title":"Lunch","amount":350.50,"category":"Food","expense_date":"10-06-2025"}`,
			headers:    userHeader(user.ID),
			wantStatus: 400,
			wantMsg:    "Expense date must be in YYYY-MM-DD format",
		},
		{
			name:       "no auth header",
			body:       `{"title":"Lunch","amount":350.50,"category":"Food","expense_date":"2025-06-10"}`,
			headers:    nil,
			wantStatus: 401,
			wantMsg:    "Unauthorized",
		},
		{
			name:       "malformed JSON",
			body:       `{bad}`,
			headers:    userHeader(user.ID),
			wantStatus: 400,
			wantMsg:    "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := execExpense(t, "Create", "/api/v1/expenses", tt.body, tt.headers, nil)
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if msg(t, rr) != tt.wantMsg {
				t.Errorf("message = %q, want %q", msg(t, rr), tt.wantMsg)
			}
		})
	}
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestListExpenses(t *testing.T) {
	defer setupExpenseCtrlTestEnv(t)()
	user := seedTestUser(t, 1, "bob@example.com")
	seedTestExpense(t, 1, user.ID, "Lunch", "Food", "2025-06-10", 200)
	seedTestExpense(t, 2, user.ID, "Bus", "Transport", "2025-06-11", 50)
	seedTestExpense(t, 3, user.ID, "Dinner", "Food", "2025-06-20", 300)

	tests := []struct {
		name       string
		url        string
		wantStatus int
		wantMsg    string
	}{
		{"list all", "/api/v1/expenses", 200, "Expenses retrieved"},
		{"filter by category", "/api/v1/expenses?category=Food", 200, "Expenses retrieved"},
		{"filter by date range", "/api/v1/expenses?date_from=2025-06-10&date_to=2025-06-11", 200, "Expenses retrieved"},
		{"with limit", "/api/v1/expenses?limit=2", 200, "Expenses retrieved"},
		{"sort by amount asc", "/api/v1/expenses?sort_by=amount&sort_order=asc", 200, "Expenses retrieved"},
		{"invalid category", "/api/v1/expenses?category=BadCat", 400, "Invalid category"},
		{"invalid sort_by", "/api/v1/expenses?sort_by=title", 400, "sort_by must be 'amount' or 'expense_date'"},
		{"invalid sort_order", "/api/v1/expenses?sort_order=random", 400, "sort_order must be 'asc' or 'desc'"},
		{"invalid limit", "/api/v1/expenses?limit=abc", 400, "limit must be a positive integer"},
		{"date_from after date_to", "/api/v1/expenses?date_from=2025-06-30&date_to=2025-06-01", 400, "date_from cannot be after date_to"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := execExpense(t, "List", tt.url, "", userHeader(user.ID), nil)
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if msg(t, rr) != tt.wantMsg {
				t.Errorf("message = %q, want %q", msg(t, rr), tt.wantMsg)
			}
		})
	}
}

// ── GetOne ────────────────────────────────────────────────────────────────────

func TestGetOneExpense(t *testing.T) {
	defer setupExpenseCtrlTestEnv(t)()
	user := seedTestUser(t, 1, "carol@example.com")
	expense := seedTestExpense(t, 1, user.ID, "Gym", "Healthcare", "2025-06-05", 100)

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantMsg    string
	}{
		{"valid get", fmt.Sprintf("%d", expense.ID), 200, "Expense retrieved"},
		{"not found", "9999", 404, "Expense not found"},
		{"invalid ID", "abc", 400, "Invalid expense ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := execExpense(t, "GetOne", "/api/v1/expenses/"+tt.id, "", userHeader(user.ID), map[string]string{":id": tt.id})
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if msg(t, rr) != tt.wantMsg {
				t.Errorf("message = %q, want %q", msg(t, rr), tt.wantMsg)
			}
		})
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUpdateExpense(t *testing.T) {
	defer setupExpenseCtrlTestEnv(t)()
	user := seedTestUser(t, 1, "dave@example.com")
	expense := seedTestExpense(t, 1, user.ID, "Coffee", "Food", "2025-06-01", 80)

	tests := []struct {
		name       string
		id         string
		body       string
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "valid update",
			id:         fmt.Sprintf("%d", expense.ID),
			body:       `{"title":"Tea","amount":50.00,"category":"Food","expense_date":"2025-06-01"}`,
			wantStatus: 200,
			wantMsg:    "Expense updated successfully",
		},
		{
			name:       "not found",
			id:         "9999",
			body:       `{"title":"Tea","amount":50.00,"category":"Food","expense_date":"2025-06-01"}`,
			wantStatus: 404,
			wantMsg:    "Expense not found",
		},
		{
			name:       "invalid category",
			id:         fmt.Sprintf("%d", expense.ID),
			body:       `{"title":"Tea","amount":50.00,"category":"Nope","expense_date":"2025-06-01"}`,
			wantStatus: 400,
			wantMsg:    "Invalid category",
		},
		{
			name:       "missing title",
			id:         fmt.Sprintf("%d", expense.ID),
			body:       `{"amount":50.00,"category":"Food","expense_date":"2025-06-01"}`,
			wantStatus: 400,
			wantMsg:    "Title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := execExpense(t, "Update", "/api/v1/expenses/"+tt.id, tt.body, userHeader(user.ID), map[string]string{":id": tt.id})
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if msg(t, rr) != tt.wantMsg {
				t.Errorf("message = %q, want %q", msg(t, rr), tt.wantMsg)
			}
		})
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDeleteExpense(t *testing.T) {
	defer setupExpenseCtrlTestEnv(t)()
	user := seedTestUser(t, 1, "eve@example.com")
	expense := seedTestExpense(t, 1, user.ID, "Movie", "Entertainment", "2025-06-15", 200)

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantMsg    string
	}{
		{"valid delete", fmt.Sprintf("%d", expense.ID), 200, "Expense deleted successfully"},
		{"already deleted", fmt.Sprintf("%d", expense.ID), 404, "Expense not found"},
		{"invalid ID", "0", 400, "Invalid expense ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := execExpense(t, "Delete", "/api/v1/expenses/"+tt.id, "", userHeader(user.ID), map[string]string{":id": tt.id})
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if msg(t, rr) != tt.wantMsg {
				t.Errorf("message = %q, want %q", msg(t, rr), tt.wantMsg)
			}
		})
	}
}

// ── Summary ───────────────────────────────────────────────────────────────────

func TestSummary(t *testing.T) {
	defer setupExpenseCtrlTestEnv(t)()
	user := seedTestUser(t, 1, "frank@example.com")
	seedTestExpense(t, 1, user.ID, "Lunch", "Food", "2025-06-10", 200)
	seedTestExpense(t, 2, user.ID, "Bus", "Transport", "2025-06-11", 50)

	tests := []struct {
		name       string
		url        string
		headers    map[string]string
		wantStatus int
		wantMsg    string
	}{
		{"valid summary", "/api/v1/expenses/summary?date_from=2025-06-01&date_to=2025-06-30", userHeader(user.ID), 200, "Summary generated"},
		{"missing date_from", "/api/v1/expenses/summary?date_to=2025-06-30", userHeader(user.ID), 400, "date_from is required"},
		{"missing date_to", "/api/v1/expenses/summary?date_from=2025-06-01", userHeader(user.ID), 400, "date_to is required"},
		{"invalid date_from format", "/api/v1/expenses/summary?date_from=01-06-2025&date_to=2025-06-30", userHeader(user.ID), 400, "date_from must be in YYYY-MM-DD format"},
		{"date_from after date_to", "/api/v1/expenses/summary?date_from=2025-06-30&date_to=2025-06-01", userHeader(user.ID), 400, "date_from cannot be after date_to"},
		{"no auth header", "/api/v1/expenses/summary?date_from=2025-06-01&date_to=2025-06-30", nil, 401, "Unauthorized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := execExpense(t, "Summary", tt.url, "", tt.headers, nil)
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if msg(t, rr) != tt.wantMsg {
				t.Errorf("message = %q, want %q", msg(t, rr), tt.wantMsg)
			}
		})
	}
}

// ── Ownership isolation ───────────────────────────────────────────────────────

func TestExpenseOwnership(t *testing.T) {
	defer setupExpenseCtrlTestEnv(t)()

	userA := seedTestUser(t, 1, "a@example.com")
	userB := seedTestUser(t, 2, "b@example.com")
	expense := seedTestExpense(t, 1, userA.ID, "Private", "Food", "2025-06-01", 100)
	id := fmt.Sprintf("%d", expense.ID)

	t.Run("user B cannot get user A expense", func(t *testing.T) {
		rr := execExpense(t, "GetOne", "/api/v1/expenses/"+id, "", userHeader(userB.ID), map[string]string{":id": id})
		if rr.Code != 404 {
			t.Errorf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("user B cannot update user A expense", func(t *testing.T) {
		body := `{"title":"Hack","amount":1,"category":"Food","expense_date":"2025-06-01"}`
		rr := execExpense(t, "Update", "/api/v1/expenses/"+id, body, userHeader(userB.ID), map[string]string{":id": id})
		if rr.Code != 404 {
			t.Errorf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("user B cannot delete user A expense", func(t *testing.T) {
		rr := execExpense(t, "Delete", "/api/v1/expenses/"+id, "", userHeader(userB.ID), map[string]string{":id": id})
		if rr.Code != 404 {
			t.Errorf("expected 404, got %d", rr.Code)
		}
	})
}

// ── validateExpenseInput (unexported) ─────────────────────────────────────────

func TestValidateExpenseInput(t *testing.T) {
	tests := []struct {
		name    string
		input   expenseInput
		wantMsg string
	}{
		{
			name:    "valid input",
			input:   expenseInput{Title: "Lunch", Amount: 100, Category: "Food", ExpenseDate: "2025-06-01"},
			wantMsg: "",
		},
		{
			name:    "empty title",
			input:   expenseInput{Amount: 100, Category: "Food", ExpenseDate: "2025-06-01"},
			wantMsg: "Title is required",
		},
		{
			name:    "zero amount",
			input:   expenseInput{Title: "Lunch", Amount: 0, Category: "Food", ExpenseDate: "2025-06-01"},
			wantMsg: "Amount must be a positive number",
		},
		{
			name:    "missing category",
			input:   expenseInput{Title: "Lunch", Amount: 100, ExpenseDate: "2025-06-01"},
			wantMsg: "Category is required",
		},
		{
			name:    "invalid category",
			input:   expenseInput{Title: "Lunch", Amount: 100, Category: "Bogus", ExpenseDate: "2025-06-01"},
			wantMsg: "Invalid category",
		},
		{
			name:    "missing date",
			input:   expenseInput{Title: "Lunch", Amount: 100, Category: "Food"},
			wantMsg: "Expense date is required",
		},
		{
			name:    "bad date format",
			input:   expenseInput{Title: "Lunch", Amount: 100, Category: "Food", ExpenseDate: "01/06/2025"},
			wantMsg: "Expense date must be in YYYY-MM-DD format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input
			got := validateExpenseInput(&input)
			if got != tt.wantMsg {
				t.Errorf("validateExpenseInput() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

// Keep beego import used across test helpers
var _ = beego.AppConfig
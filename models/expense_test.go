package models

import (
	"os"
	"path/filepath"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
)

// setupExpenseTestEnv points models at temp CSVs and returns cleanup func.
func setupExpenseTestEnv(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	userCSV := filepath.Join(dir, "users.csv")
	expenseCSV := filepath.Join(dir, "expenses.csv")
	beego.AppConfig.Set("userscsv", userCSV)
	beego.AppConfig.Set("expensescsv", expenseCSV)
	return func() {
		os.Remove(userCSV)
		os.Remove(expenseCSV)
	}
}

// makeExpense is a helper to build a test Expense with sensible defaults.
func makeExpense(id, userID int, title, category, date string, amount float64) Expense {
	return Expense{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Amount:      amount,
		Category:    category,
		Note:        "",
		ExpenseDate: date,
	}
}

func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		category string
		want     bool
	}{
		{"Food", true},
		{"Transport", true},
		{"Housing", true},
		{"Entertainment", true},
		{"Shopping", true},
		{"Healthcare", true},
		{"Education", true},
		{"Utilities", true},
		{"Other", true},
		{"food", false},    // case sensitive
		{"FOOD", false},    // case sensitive
		{"Invalid", false}, // not in list
		{"", false},        // empty string
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := IsValidCategory(tt.category)
			if got != tt.want {
				t.Errorf("IsValidCategory(%q) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

func TestCreateExpense(t *testing.T) {
	defer setupExpenseTestEnv(t)()

	tests := []struct {
		name    string
		expense *Expense
		wantErr bool
	}{
		{
			name:    "valid Food expense",
			expense: &Expense{ID: 1, UserID: 1, Title: "Lunch", Amount: 350.50, Category: "Food", ExpenseDate: "2025-06-10"},
		},
		{
			name:    "valid Transport expense",
			expense: &Expense{ID: 2, UserID: 1, Title: "Bus", Amount: 50.00, Category: "Transport", ExpenseDate: "2025-06-11"},
		},
		{
			name:    "expense with note",
			expense: &Expense{ID: 3, UserID: 2, Title: "Dinner", Amount: 200.00, Category: "Food", Note: "With family", ExpenseDate: "2025-06-12"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateExpense(tt.expense)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateExpense() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// All 3 expenses must be persisted
	all, err := readAllExpenses()
	if err != nil {
		t.Fatalf("readAllExpenses() error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 expenses, got %d", len(all))
	}
}

func TestGetExpensesByUserID(t *testing.T) {
	defer setupExpenseTestEnv(t)()

	// Seed expenses for two different users
	_ = CreateExpense(&Expense{ID: 1, UserID: 1, Title: "Lunch", Amount: 100, Category: "Food", ExpenseDate: "2025-06-01"})
	_ = CreateExpense(&Expense{ID: 2, UserID: 1, Title: "Bus", Amount: 50, Category: "Transport", ExpenseDate: "2025-06-02"})
	_ = CreateExpense(&Expense{ID: 3, UserID: 2, Title: "Rent", Amount: 5000, Category: "Housing", ExpenseDate: "2025-06-01"})

	tests := []struct {
		name      string
		userID    int
		wantCount int
	}{
		{
			name:      "user 1 has 2 expenses",
			userID:    1,
			wantCount: 2,
		},
		{
			name:      "user 2 has 1 expense",
			userID:    2,
			wantCount: 1,
		},
		{
			name:      "user 3 has no expenses",
			userID:    3,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenses, err := GetExpensesByUserID(tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(expenses) != tt.wantCount {
				t.Errorf("count = %d, want %d", len(expenses), tt.wantCount)
			}
		})
	}
}

func TestGetExpenseByID(t *testing.T) {
	defer setupExpenseTestEnv(t)()

	_ = CreateExpense(&Expense{ID: 1, UserID: 1, Title: "Lunch", Amount: 100, Category: "Food", ExpenseDate: "2025-06-01"})
	_ = CreateExpense(&Expense{ID: 2, UserID: 2, Title: "Rent", Amount: 5000, Category: "Housing", ExpenseDate: "2025-06-01"})

	tests := []struct {
		name      string
		id        int
		userID    int
		wantFound bool
		wantTitle string
	}{
		{
			name:      "found own expense",
			id:        1,
			userID:    1,
			wantFound: true,
			wantTitle: "Lunch",
		},
		{
			name:      "expense belongs to other user",
			id:        2,
			userID:    1,
			wantFound: false,
		},
		{
			name:      "non-existing ID",
			id:        999,
			userID:    1,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense, err := GetExpenseByID(tt.id, tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if (expense != nil) != tt.wantFound {
				t.Errorf("found = %v, wantFound = %v", expense != nil, tt.wantFound)
			}
			if tt.wantFound && expense.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", expense.Title, tt.wantTitle)
			}
		})
	}
}

func TestUpdateExpense(t *testing.T) {
	defer setupExpenseTestEnv(t)()

	_ = CreateExpense(&Expense{ID: 1, UserID: 1, Title: "Lunch", Amount: 100, Category: "Food", ExpenseDate: "2025-06-01"})

	tests := []struct {
		name    string
		updated *Expense
		wantErr bool
	}{
		{
			name:    "valid update",
			updated: &Expense{ID: 1, UserID: 1, Title: "Dinner", Amount: 200, Category: "Food", ExpenseDate: "2025-06-01"},
			wantErr: false,
		},
		{
			name:    "non-existing expense",
			updated: &Expense{ID: 999, UserID: 1, Title: "Ghost", Amount: 100, Category: "Food", ExpenseDate: "2025-06-01"},
			wantErr: true,
		},
		{
			name:    "wrong user ID",
			updated: &Expense{ID: 1, UserID: 99, Title: "Hack", Amount: 100, Category: "Food", ExpenseDate: "2025-06-01"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateExpense(tt.updated)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateExpense() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Confirm the title actually changed
	e, _ := GetExpenseByID(1, 1)
	if e == nil {
		t.Fatal("expense 1 not found after update")
	}
	if e.Title != "Dinner" {
		t.Errorf("title after update = %q, want Dinner", e.Title)
	}
}

func TestDeleteExpense(t *testing.T) {
	defer setupExpenseTestEnv(t)()

	_ = CreateExpense(&Expense{ID: 1, UserID: 1, Title: "Lunch", Amount: 100, Category: "Food", ExpenseDate: "2025-06-01"})
	_ = CreateExpense(&Expense{ID: 2, UserID: 1, Title: "Bus", Amount: 50, Category: "Transport", ExpenseDate: "2025-06-02"})

	tests := []struct {
		name      string
		id        int
		userID    int
		wantErr   bool
		wantCount int // remaining expenses for userID 1
	}{
		{
			name:      "valid delete",
			id:        1,
			userID:    1,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "already deleted",
			id:        1,
			userID:    1,
			wantErr:   true,
			wantCount: 1,
		},
		{
			name:      "wrong user cannot delete",
			id:        2,
			userID:    99,
			wantErr:   true,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteExpense(tt.id, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteExpense() error = %v, wantErr %v", err, tt.wantErr)
			}
			remaining, _ := GetExpensesByUserID(1)
			if len(remaining) != tt.wantCount {
				t.Errorf("remaining count = %d, want %d", len(remaining), tt.wantCount)
			}
		})
	}
}

func TestGetNextExpenseID(t *testing.T) {
	defer setupExpenseTestEnv(t)()

	// Empty CSV → first ID is 1
	got := GetNextExpenseID()
	if got != 1 {
		t.Errorf("empty: GetNextExpenseID() = %d, want 1", got)
	}

	_ = CreateExpense(&Expense{ID: 1, UserID: 1, Title: "A", Amount: 10, Category: "Food", ExpenseDate: "2025-06-01"})
	_ = CreateExpense(&Expense{ID: 2, UserID: 1, Title: "B", Amount: 20, Category: "Food", ExpenseDate: "2025-06-02"})

	got = GetNextExpenseID()
	if got != 3 {
		t.Errorf("after 2 expenses: GetNextExpenseID() = %d, want 3", got)
	}
}

// TestRecordToExpense tests the unexported CSV parsing function directly.
func TestRecordToExpense(t *testing.T) {
	tests := []struct {
		name    string
		record  []string
		wantErr bool
		wantID  int
	}{
		{
			name:    "valid record",
			record:  []string{"1", "1", "Lunch", "350.50", "Food", "Team lunch", "2025-06-10", "2025-06-10T14:30:00Z"},
			wantErr: false,
			wantID:  1,
		},
		{
			name:    "too few fields",
			record:  []string{"1", "1", "Lunch"},
			wantErr: true,
		},
		{
			name:    "invalid ID",
			record:  []string{"abc", "1", "Lunch", "350.50", "Food", "", "2025-06-10", "2025-06-10T14:30:00Z"},
			wantErr: true,
		},
		{
			name:    "invalid amount",
			record:  []string{"1", "1", "Lunch", "notanumber", "Food", "", "2025-06-10", "2025-06-10T14:30:00Z"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense, err := recordToExpense(tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("recordToExpense() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && expense.ID != tt.wantID {
				t.Errorf("ID = %d, want %d", expense.ID, tt.wantID)
			}
		})
	}
}

// TestExpenseToRecord tests the unexported CSV serialisation function directly.
func TestExpenseToRecord(t *testing.T) {
	e := &Expense{
		ID: 1, UserID: 1, Title: "Lunch", Amount: 350.50,
		Category: "Food", Note: "team", ExpenseDate: "2025-06-10",
		CreatedAt: "2025-06-10T14:30:00Z",
	}
	record := expenseToRecord(e)

	if len(record) != 8 {
		t.Errorf("record length = %d, want 8", len(record))
	}
	if record[0] != "1" {
		t.Errorf("id = %q, want 1", record[0])
	}
	if record[3] != "350.50" {
		t.Errorf("amount = %q, want 350.50", record[3])
	}
}
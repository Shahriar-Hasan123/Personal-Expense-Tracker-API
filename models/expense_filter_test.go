package models

import "testing"

func TestFilterByCategory(t *testing.T) {
	expenses := []Expense{
		{ID: 1, Category: "Food"},
		{ID: 2, Category: "Transport"},
		{ID: 3, Category: "Food"},
	}

	tests := []struct {
		name      string
		category  string
		wantCount int
	}{
		{"filter Food", "Food", 2},
		{"filter Transport", "Transport", 1},
		{"filter Housing returns empty", "Housing", 0},
		{"empty category returns all", "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterByCategory(expenses, tt.category)
			if len(got) != tt.wantCount {
				t.Errorf("count = %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestFilterByDateRange(t *testing.T) {
	expenses := []Expense{
		{ID: 1, ExpenseDate: "2025-06-01"},
		{ID: 2, ExpenseDate: "2025-06-10"},
		{ID: 3, ExpenseDate: "2025-06-20"},
		{ID: 4, ExpenseDate: "2025-06-30"},
	}

	tests := []struct {
		name      string
		dateFrom  string
		dateTo    string
		wantCount int
	}{
		{"full range", "2025-06-01", "2025-06-30", 4},
		{"mid range", "2025-06-10", "2025-06-20", 2},
		{"only date_from", "2025-06-15", "", 2},
		{"only date_to", "", "2025-06-10", 2},
		{"no bounds returns all", "", "", 4},
		{"no results in range", "2025-07-01", "2025-07-31", 0},
		{"single day", "2025-06-10", "2025-06-10", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterByDateRange(expenses, tt.dateFrom, tt.dateTo)
			if len(got) != tt.wantCount {
				t.Errorf("count = %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestSortExpenses(t *testing.T) {
	expenses := []Expense{
		{ID: 1, Title: "Bus", Amount: 50, ExpenseDate: "2025-06-10"},
		{ID: 2, Title: "Rent", Amount: 5000, ExpenseDate: "2025-06-01"},
		{ID: 3, Title: "Lunch", Amount: 200, ExpenseDate: "2025-06-20"},
	}

	tests := []struct {
		name       string
		sortBy     string
		sortOrder  string
		wantFirst  string
		wantLast   string
	}{
		{"amount asc", "amount", "asc", "Bus", "Rent"},
		{"amount desc", "amount", "desc", "Rent", "Bus"},
		{"date asc", "expense_date", "asc", "Rent", "Lunch"},
		{"date desc", "expense_date", "desc", "Lunch", "Rent"},
		{"default (no params) → date desc", "", "", "Lunch", "Rent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy slice so each test is independent
			input := make([]Expense, len(expenses))
			copy(input, expenses)

			got := sortExpenses(input, tt.sortBy, tt.sortOrder)
			if got[0].Title != tt.wantFirst {
				t.Errorf("first = %q, want %q", got[0].Title, tt.wantFirst)
			}
			if got[len(got)-1].Title != tt.wantLast {
				t.Errorf("last = %q, want %q", got[len(got)-1].Title, tt.wantLast)
			}
		})
	}
}

func TestApplyFilters(t *testing.T) {
	expenses := []Expense{
		{ID: 1, Title: "Lunch", Amount: 200, Category: "Food", ExpenseDate: "2025-06-10"},
		{ID: 2, Title: "Bus", Amount: 50, Category: "Transport", ExpenseDate: "2025-06-11"},
		{ID: 3, Title: "Dinner", Amount: 300, Category: "Food", ExpenseDate: "2025-06-20"},
		{ID: 4, Title: "Rent", Amount: 5000, Category: "Housing", ExpenseDate: "2025-06-01"},
	}

	tests := []struct {
		name      string
		opts      FilterOptions
		wantCount int
		wantFirst string
	}{
		{
			name:      "no filters returns all",
			opts:      FilterOptions{},
			wantCount: 4,
		},
		{
			name:      "category filter",
			opts:      FilterOptions{Category: "Food"},
			wantCount: 2,
		},
		{
			name:      "date range filter",
			opts:      FilterOptions{DateFrom: "2025-06-10", DateTo: "2025-06-11"},
			wantCount: 2,
		},
		{
			name:      "limit",
			opts:      FilterOptions{Limit: 2},
			wantCount: 2,
		},
		{
			name:      "sort amount asc — first is Bus",
			opts:      FilterOptions{SortBy: "amount", SortOrder: "asc"},
			wantCount: 4,
			wantFirst: "Bus",
		},
		{
			name:      "category + date combined",
			opts:      FilterOptions{Category: "Food", DateFrom: "2025-06-10", DateTo: "2025-06-15"},
			wantCount: 1,
		},
		{
			name:      "empty result for out-of-range date",
			opts:      FilterOptions{DateFrom: "2025-07-01", DateTo: "2025-07-31"},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := make([]Expense, len(expenses))
			copy(input, expenses)

			got := ApplyFilters(input, tt.opts)
			if len(got) != tt.wantCount {
				t.Errorf("count = %d, want %d", len(got), tt.wantCount)
			}
			if tt.wantFirst != "" && len(got) > 0 && got[0].Title != tt.wantFirst {
				t.Errorf("first = %q, want %q", got[0].Title, tt.wantFirst)
			}
		})
	}
}
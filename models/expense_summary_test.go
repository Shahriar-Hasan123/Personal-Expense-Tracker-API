package models

import "testing"

func TestBuildSummary(t *testing.T) {
	tests := []struct {
		name            string
		expenses        []Expense
		dateFrom        string
		dateTo          string
		wantTotal       float64
		wantCount       int
		wantCatCount    int
		wantFirstCat    string
		wantFirstTotal  float64
		wantFirstCount  int
	}{
		{
			name: "multiple categories",
			expenses: []Expense{
				{ID: 1, Category: "Food", Amount: 200},
				{ID: 2, Category: "Food", Amount: 300},
				{ID: 3, Category: "Transport", Amount: 50},
			},
			dateFrom:       "2025-06-01",
			dateTo:         "2025-06-30",
			wantTotal:      550,
			wantCount:      3,
			wantCatCount:   2,
			wantFirstCat:   "Food",   // AllowedCategories order
			wantFirstTotal: 500,
			wantFirstCount: 2,
		},
		{
			name:         "empty expense list",
			expenses:     []Expense{},
			dateFrom:     "2025-06-01",
			dateTo:       "2025-06-30",
			wantTotal:    0,
			wantCount:    0,
			wantCatCount: 0,
		},
		{
			name: "single expense",
			expenses: []Expense{
				{ID: 1, Category: "Healthcare", Amount: 1500},
			},
			dateFrom:       "2025-06-01",
			dateTo:         "2025-06-30",
			wantTotal:      1500,
			wantCount:      1,
			wantCatCount:   1,
			wantFirstCat:   "Healthcare",
			wantFirstTotal: 1500,
			wantFirstCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := BuildSummary(tt.expenses, tt.dateFrom, tt.dateTo)

			if summary.TotalAmount != tt.wantTotal {
				t.Errorf("TotalAmount = %.2f, want %.2f", summary.TotalAmount, tt.wantTotal)
			}
			if summary.TotalCount != tt.wantCount {
				t.Errorf("TotalCount = %d, want %d", summary.TotalCount, tt.wantCount)
			}
			if len(summary.ByCategory) != tt.wantCatCount {
				t.Errorf("ByCategory length = %d, want %d", len(summary.ByCategory), tt.wantCatCount)
			}
			if tt.wantCatCount > 0 {
				first := summary.ByCategory[0]
				if first.Category != tt.wantFirstCat {
					t.Errorf("first category = %q, want %q", first.Category, tt.wantFirstCat)
				}
				if first.Total != tt.wantFirstTotal {
					t.Errorf("first total = %.2f, want %.2f", first.Total, tt.wantFirstTotal)
				}
				if first.Count != tt.wantFirstCount {
					t.Errorf("first count = %d, want %d", first.Count, tt.wantFirstCount)
				}
			}
			if summary.DateFrom != tt.dateFrom {
				t.Errorf("DateFrom = %q, want %q", summary.DateFrom, tt.dateFrom)
			}
			if summary.DateTo != tt.dateTo {
				t.Errorf("DateTo = %q, want %q", summary.DateTo, tt.dateTo)
			}
		})
	}
}

func TestRoundFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{350.5, 350.5},
		{350.555, 350.56},
		{0.0, 0.0},
		{1000.004, 1000.0},
		{99.999, 100.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := roundFloat(tt.input)
			if got != tt.want {
				t.Errorf("roundFloat(%.4f) = %.4f, want %.4f", tt.input, got, tt.want)
			}
		})
	}
}
package models

// CategorySummary holds the total and count for a single category.
type CategorySummary struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
	Count    int     `json:"count"`
}

// Summary holds the full spending summary for a date range.
type Summary struct {
	DateFrom    string            `json:"date_from"`
	DateTo      string            `json:"date_to"`
	TotalAmount float64           `json:"total_amount"`
	TotalCount  int               `json:"total_count"`
	ByCategory  []CategorySummary `json:"by_category"`
}

// BuildSummary computes totals and per-category breakdown for the given expenses.
// The expenses slice must already be filtered to the desired date range.
func BuildSummary(expenses []Expense, dateFrom, dateTo string) Summary {
	totalAmount := 0.0
	// Use a map to accumulate totals per category
	categoryTotals := make(map[string]float64)
	categoryCounts := make(map[string]int)

	for _, e := range expenses {
		totalAmount += e.Amount
		categoryTotals[e.Category] += e.Amount
		categoryCounts[e.Category]++
	}

	// Build the by_category slice in AllowedCategories order
	// so the response is deterministic (not random map iteration order)
	byCategory := make([]CategorySummary, 0)
	for _, cat := range AllowedCategories {
		if total, ok := categoryTotals[cat]; ok {
			byCategory = append(byCategory, CategorySummary{
				Category: cat,
				Total:    roundFloat(total),
				Count:    categoryCounts[cat],
			})
		}
	}

	return Summary{
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		TotalAmount: roundFloat(totalAmount),
		TotalCount:  len(expenses),
		ByCategory:  byCategory,
	}
}

// roundFloat rounds a float64 to 2 decimal places to avoid floating point noise.
func roundFloat(val float64) float64 {
	// Multiply, truncate, divide — no external library needed
	return float64(int(val*100+0.5)) / 100
}
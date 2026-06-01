package models

import (
	"sort"
)

// FilterOptions holds all supported query parameters for listing expenses.
type FilterOptions struct {
	Category  string
	DateFrom  string
	DateTo    string
	SortBy    string
	SortOrder string
	Limit     int
}

// ApplyFilters filters, sorts, and limits a slice of expenses.
// Steps follow the PDF exactly: filter category → filter date → sort → limit.
func ApplyFilters(expenses []Expense, opts FilterOptions) []Expense {
	result := filterByCategory(expenses, opts.Category)
	result = filterByDateRange(result, opts.DateFrom, opts.DateTo)
	result = sortExpenses(result, opts.SortBy, opts.SortOrder)

	if opts.Limit > 0 && opts.Limit < len(result) {
		result = result[:opts.Limit]
	}
	return result
}

// filterByCategory keeps only expenses matching the given category.
// If category is empty, all expenses are returned unchanged.
func filterByCategory(expenses []Expense, category string) []Expense {
	if category == "" {
		return expenses
	}
	result := make([]Expense, 0)
	for _, e := range expenses {
		if e.Category == category {
			result = append(result, e)
		}
	}
	return result
}

// filterByDateRange keeps expenses whose ExpenseDate falls within
// [dateFrom, dateTo] inclusive. Empty strings mean no bound on that side.
// YYYY-MM-DD strings compare correctly as plain strings in Go.
func filterByDateRange(expenses []Expense, dateFrom, dateTo string) []Expense {
	if dateFrom == "" && dateTo == "" {
		return expenses
	}
	result := make([]Expense, 0)
	for _, e := range expenses {
		if dateFrom != "" && e.ExpenseDate < dateFrom {
			continue
		}
		if dateTo != "" && e.ExpenseDate > dateTo {
			continue
		}
		result = append(result, e)
	}
	return result
}

// sortExpenses sorts the expenses slice by the given field and order.
// Default sort: expense_date desc.
func sortExpenses(expenses []Expense, sortBy, sortOrder string) []Expense {
	if sortBy == "" {
		sortBy = "expense_date"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	sort.Slice(expenses, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "amount":
			less = expenses[i].Amount < expenses[j].Amount
		default: // expense_date
			less = expenses[i].ExpenseDate < expenses[j].ExpenseDate
		}

		if sortOrder == "asc" {
			return less
		}
		return !less // desc
	})

	return expenses
}
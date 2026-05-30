package models

import (
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

// Expense represents a single expense record.
type Expense struct {
	ID          int
	UserID      int
	Title       string
	Amount      float64
	Category    string
	Note        string
	ExpenseDate string
	CreatedAt   string
}

// AllowedCategories is the list of valid expense categories.
var AllowedCategories = []string{
	"Food", "Transport", "Housing", "Entertainment",
	"Shopping", "Healthcare", "Education", "Utilities", "Other",
}

// IsValidCategory checks if the given category is in the allowed list.
func IsValidCategory(category string) bool {
	for _, c := range AllowedCategories {
		if c == category {
			return true
		}
	}
	return false
}

// expensesFilePath returns the path to the expenses CSV file from config.
func expensesFilePath() string {
	path, _ := beego.AppConfig.String("expensescsv")
	if path == "" {
		return "data/expenses.csv"
	}
	return path
}

// expenseToRecord converts an Expense struct to a CSV string slice.
func expenseToRecord(e *Expense) []string {
	return []string{
		strconv.Itoa(e.ID),
		strconv.Itoa(e.UserID),
		e.Title,
		strconv.FormatFloat(e.Amount, 'f', 2, 64),
		e.Category,
		e.Note,
		e.ExpenseDate,
		e.CreatedAt,
	}
}

// recordToExpense converts a CSV string slice to an Expense struct.
func recordToExpense(record []string) (*Expense, error) {
	if len(record) < 8 {
		return nil, errors.New("invalid expense record")
	}
	id, err := strconv.Atoi(record[0])
	if err != nil {
		return nil, err
	}
	userID, err := strconv.Atoi(record[1])
	if err != nil {
		return nil, err
	}
	amount, err := strconv.ParseFloat(record[3], 64)
	if err != nil {
		return nil, err
	}
	return &Expense{
		ID:          id,
		UserID:      userID,
		Title:       record[2],
		Amount:      amount,
		Category:    record[4],
		Note:        record[5],
		ExpenseDate: record[6],
		CreatedAt:   record[7],
	}, nil
}

// readAllExpenses reads every expense row from the CSV file.
func readAllExpenses() ([]Expense, error) {
	path := expensesFilePath()
	if err := ensureFile(path); err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var expenses []Expense
	for i, record := range records {
		// Skip header row
		if i == 0 && record[0] == "id" {
			continue
		}
		e, err := recordToExpense(record)
		if err != nil {
			logs.Warn("Skipping malformed expense record: %v", err)
			continue
		}
		expenses = append(expenses, *e)
	}
	return expenses, nil
}

// writeAllExpenses rewrites the entire expenses CSV with the given slice.
func writeAllExpenses(expenses []Expense) error {
	path := expensesFilePath()

	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Always write header first
	if err := writer.Write([]string{
		"id", "user_id", "title", "amount",
		"category", "note", "expense_date", "created_at",
	}); err != nil {
		return err
	}

	for _, e := range expenses {
		if err := writer.Write(expenseToRecord(&e)); err != nil {
			return err
		}
	}
	return nil
}

// GetNextExpenseID determines the next available expense ID.
func GetNextExpenseID() int {
	expenses, err := readAllExpenses()
	if err != nil || len(expenses) == 0 {
		return 1
	}
	max := 0
	for _, e := range expenses {
		if e.ID > max {
			max = e.ID
		}
	}
	return max + 1
}

// GetExpensesByUserID returns all expenses belonging to a specific user.
func GetExpensesByUserID(userID int) ([]Expense, error) {
	all, err := readAllExpenses()
	if err != nil {
		return nil, err
	}
	var result []Expense
	for _, e := range all {
		if e.UserID == userID {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetExpenseByID returns a single expense by ID, scoped to the user.
func GetExpenseByID(id int, userID int) (*Expense, error) {
	all, err := readAllExpenses()
	if err != nil {
		return nil, err
	}
	for _, e := range all {
		if e.ID == id && e.UserID == userID {
			return &e, nil
		}
	}
	return nil, nil
}

// CreateExpense appends a new expense to the CSV file.
func CreateExpense(expense *Expense) error {
	path := expensesFilePath()
	if err := ensureFile(path); err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	writeHeader := info.Size() == 0

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	if writeHeader {
		if err := writer.Write([]string{
			"id", "user_id", "title", "amount",
			"category", "note", "expense_date", "created_at",
		}); err != nil {
			return err
		}
	}

	expense.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	return writer.Write(expenseToRecord(expense))
}

// UpdateExpense rewrites the CSV replacing the matching expense row.
func UpdateExpense(updated *Expense) error {
	all, err := readAllExpenses()
	if err != nil {
		return err
	}

	found := false
	for i, e := range all {
		if e.ID == updated.ID && e.UserID == updated.UserID {
			// Preserve original created_at
			updated.CreatedAt = e.CreatedAt
			all[i] = *updated
			found = true
			break
		}
	}

	if !found {
		return errors.New("expense not found")
	}

	return writeAllExpenses(all)
}

// DeleteExpense rewrites the CSV omitting the matching expense row.
func DeleteExpense(id int, userID int) error {
	all, err := readAllExpenses()
	if err != nil {
		return err
	}

	filtered := make([]Expense, 0, len(all))
	found := false
	for _, e := range all {
		if e.ID == id && e.UserID == userID {
			found = true
			continue // skip this row
		}
		filtered = append(filtered, e)
	}

	if !found {
		return errors.New("expense not found")
	}

	return writeAllExpenses(filtered)
}

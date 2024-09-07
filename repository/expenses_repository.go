package repository

import (
	"go-sheet/model"

	"errors"

	"github.com/nedpals/supabase-go"
)

type ExpensesRepository struct {
	connection *supabase.Client
}

func NewExpensesRepository(connection *supabase.Client) *ExpensesRepository {
	return &ExpensesRepository{
		connection: connection,
	}
}

func (r *ExpensesRepository) GetAll() ([]model.Expenses, error) {
	var expenses []model.Expenses
	if r.connection == nil || r.connection.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}
	err := r.connection.DB.From("expenses").Select("*").Execute(&expenses)
	if err != nil {
		return nil, err
	}

	return expenses, nil
}

func (r *ExpensesRepository) CreateExpense(expense model.Expenses) (model.Expenses, error) {
	var result model.Expenses
	err := r.connection.DB.From("expenses").Insert(map[string]interface{}{
		"name":       expense.Name,
		"planned":    expense.Planned,
		"spent":      expense.Spent,
		"difference": expense.Difference,
	}).Execute(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

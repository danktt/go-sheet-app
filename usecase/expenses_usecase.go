package usecase

import (
	"go-sheet/model"
	"go-sheet/repository"
)

type ExpensesUseCase struct {
	expensesRepository *repository.ExpensesRepository
}

func NewExpensesUseCase(expensesRepository *repository.ExpensesRepository) *ExpensesUseCase {
	return &ExpensesUseCase{
		expensesRepository: expensesRepository,
	}
}

func (u *ExpensesUseCase) GetAll() ([]model.Expenses, error) {
	return u.expensesRepository.GetAll()
}
func (u *ExpensesUseCase) CreateExpense(expense model.Expenses) (model.Expenses, error) {
	return u.expensesRepository.CreateExpense(expense)
}

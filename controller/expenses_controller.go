package controller

import (
	"go-sheet/model"
	"go-sheet/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type expensesController struct {
	ExpensesUseCase *usecase.ExpensesUseCase
}

func NewExpensesController(usecase *usecase.ExpensesUseCase) *expensesController {
	return &expensesController{
		ExpensesUseCase: usecase,
	}
}

func (c *expensesController) GetAll(ctx *gin.Context) {
	expenses, err := c.ExpensesUseCase.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, expenses)
}

func (c *expensesController) CreateExpense(ctx *gin.Context) {
	var expense model.Expenses
	if err := ctx.ShouldBindJSON(&expense); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createdExpense, err := c.ExpensesUseCase.CreateExpense(expense)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, createdExpense)
}

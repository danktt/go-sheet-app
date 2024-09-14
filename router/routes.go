package routes

import (
	handlersAnalytic "go-sheet/handlers/analytic"
	handlersCategories "go-sheet/handlers/categories"
	handlersExpenses "go-sheet/handlers/expenses"
	handlersPaidType "go-sheet/handlers/paid_type"
	handlersStatus "go-sheet/handlers/status"

	"github.com/gin-gonic/gin"
)

func InitializeRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/expenses", handlersExpenses.ListMonthlyExpenses)
		v1.POST("/expenses", handlersExpenses.CreateExpense)
		// v1.GET("/expenses/:id", handlersExpenses.ShowExpense)
		// v1.PUT("/expenses/:id", handlersExpenses.UpdateExpense)
		// v1.DELETE("/expenses/:id", handlersExpenses.DeleteExpense)

		// Categories
		v1.GET("/categories", handlersCategories.GetCategories)
		v1.POST("/categories", handlersCategories.CreateCategory)
		v1.DELETE("/categories/:id", handlersCategories.DeleteCategory)
		v1.PUT("/categories/:id", handlersCategories.UpdateCategory)
		// Paid Types
		v1.GET("/paid-types", handlersPaidType.ListPaidTypes)
		v1.POST("/paid-types", handlersPaidType.CreatePaidType)

		// Status
		v1.GET("/status", handlersStatus.ListStatus)
		v1.POST("/status", handlersStatus.CreateStatus)
		v1.DELETE("/status/:id", handlersStatus.DeleteStatus)

		// Analytic
		v1.GET("/dashboard/analytic/total", handlersAnalytic.GetAnalyticTotal)
		v1.GET("/dashboard/analytic/pending-payments", handlersAnalytic.GetPendingPayment)
	}

}

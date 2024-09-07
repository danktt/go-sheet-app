package routes

import (
	"go-sheet/handlers"

	"github.com/gin-gonic/gin"
)

func InitializeRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/expenses", handlers.ListExpenses)
		v1.POST("/expenses", handlers.CreateExpense)
		v1.GET("/expenses/:id", handlers.ShowExpense)
		v1.PUT("/expenses/:id", handlers.UpdateExpense)
		v1.DELETE("/expenses/:id", handlers.DeleteExpense)
	}

}

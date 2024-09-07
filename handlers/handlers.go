package handlers

import "github.com/gin-gonic/gin"

func CreateExpense(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Successfully created expense",
	})
	ctx.AbortWithStatus(200)
}

func ShowExpense(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Successfully created expense",
	})
	ctx.AbortWithStatus(200)
}

func DeleteExpense(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Successfully delete expense",
	})
	ctx.AbortWithStatus(200)
}

func UpdateExpense(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Successfully update expense",
	})
	ctx.AbortWithStatus(200)
}

func ListExpenses(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Successfully list expenses",
	})
	ctx.AbortWithStatus(200)
}

package handlers

import (
	"go-sheet/db"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Expenses struct {
	ID          string  `json:"uuid"`
	Created_at  string  `json:"createdAt"`
	Updated_at  string  `json:"updatedAt"`
	Description string  `json:"description"`
	Planned     float64 `json:"planned"`
	Spent       float64 `json:"spent"`
	Difference  float64 `json:"difference"`
	Category    string  `json:"category"`
	Paid_at     *string `json:"paidAt"`
	Paid_by     string  `json:"paidBy"`
}

func CreateExpense(ctx *gin.Context) {
	var expense Expenses

	if err := ctx.ShouldBindJSON(&expense); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	expense.ID = uuid.NewString()
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	expense.Created_at = currentTime
	expense.Updated_at = currentTime
	expense.Difference = expense.Planned - expense.Spent

	conn, err := db.OpenConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error connecting to database",
			"error":   err.Error(),
		})
		return
	}
	defer conn.Close()

	sql := `INSERT INTO expenses (id, created_at, updated_at, description, planned, spent, difference, category, paid_at, paid_by) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = conn.Exec(sql, expense.ID, expense.Created_at, expense.Updated_at, expense.Description, expense.Planned, expense.Spent, expense.Difference, expense.Category, expense.Paid_at, expense.Paid_by)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error inserting data into database",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Successfully created expense",
		"expense": expense,
	})
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
	// Parâmetros de paginação
	page := ctx.DefaultQuery("page", "1")
	limit := ctx.DefaultQuery("limit", "10")

	// Converter para inteiros
	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	// Calcular o offset
	offset := (pageInt - 1) * limitInt

	conn, err := db.OpenConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error connecting to database",
			"error":   err.Error(),
		})
		return
	}
	defer conn.Close()

	// Consulta paginada
	rows, err := conn.Query("SELECT * FROM expenses ORDER BY created_at DESC LIMIT $1 OFFSET $2", limitInt, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error querying database",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	var expenses []Expenses

	for rows.Next() {
		var expense Expenses
		err := rows.Scan(
			&expense.ID,
			&expense.Created_at,
			&expense.Updated_at,
			&expense.Description,
			&expense.Planned,
			&expense.Spent,
			&expense.Difference,
			&expense.Category,
			&expense.Paid_at,
			&expense.Paid_by,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error scanning database result",
				"error":   err.Error(),
			})
			return
		}
		expenses = append(expenses, expense)
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error after scanning database results",
			"error":   err.Error(),
		})
		return
	}

	// Contar o total de registros
	var total int
	err = conn.QueryRow("SELECT COUNT(*) FROM expenses").Scan(&total)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error counting total records",
			"error":   err.Error(),
		})
		return
	}

	// Calcular o total de páginas
	totalPages := (total + limitInt - 1) / limitInt

	ctx.JSON(http.StatusOK, gin.H{
		"data":        expenses,
		"currentPage": pageInt,
		"perPage":     limitInt,
		"total":       total,
		"totalPages":  totalPages,
	})
}

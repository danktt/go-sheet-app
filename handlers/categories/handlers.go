package categories

import (
	"go-sheet/db"
	"net/http"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Category struct {
	ID            string  `json:"uuid" `
	Name          string  `json:"name" binding:"required"`
	PlannedAmount float64 `json:"plannedAmount" binding:"required"`
	Color         string  `json:"color"`
	Description   string  `json:"description"`
}

type CategoryResponse struct {
	CategoryID     string         `json:"categoryId"`
	CategoryName   string         `json:"categoryName"`
	PlannedAmount  float64        `json:"plannedAmount"`
	Color          string         `json:"color"`
	ReferenceMonth sql.NullString `json:"referenceMonth"`
}

func GetCategories(ctx *gin.Context) {
	conn, err := db.OpenConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error connecting to database",
			"error":   err.Error(),
		})
		return
	}
	defer conn.Close()

	sqlQuery := `SELECT category_id, category_name, amount_planned, category_color, reference_month FROM categories`

	rows, err := conn.Query(sqlQuery)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error querying database",
			"error":   err.Error(),
			"rows":    err,
		})
		return
	}
	defer rows.Close()

	var categories []CategoryResponse

	for rows.Next() {
		var category CategoryResponse
		var referenceMonth sql.NullString
		err := rows.Scan(
			&category.CategoryID,
			&category.CategoryName,
			&category.PlannedAmount,
			&category.Color,
			&referenceMonth,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Error scanning database result",
				"error":   err.Error(),
			})
			return
		}

		category.ReferenceMonth = referenceMonth
		categories = append(categories, category)
	}
	if len(categories) == 0 {
		categories = []CategoryResponse{}
	}
	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error after scanning database results",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Successfully retrieved categories",
		"data":    categories,
	})
}

func CreateCategory(ctx *gin.Context) {
	var category Category

	if err := ctx.ShouldBindJSON(&category); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	conn, err := db.OpenConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer conn.Close()

	if conn == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection is nil"})
		return
	}

	// Inserir categoria na tabela de categorias
	sqlQuery := `INSERT INTO categories (category_id, category_name, amount_planned, category_color) 
		VALUES ($1, $2, $3, $4)`
	category.ID = uuid.NewString()
	_, err = conn.Exec(sqlQuery, category.ID, category.Name, category.PlannedAmount, category.Color)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data into database"})
		return
	}

	// Modificar o formato da data para ser compatível com o tipo date do PostgreSQL
	referenceMonth := time.Now().Format("2006-01-01") // Alterado para incluir o dia

	// Atualizar a query para incluir reference_month, description e status_id
	monthlyExpenseQuery := `INSERT INTO monthly_expenses (category_id, reference_month, spent_amount, amount_planned, difference_amount, payment_date, file, description, status_id) 
		VALUES ($1, $2, NULL, $3, NULL, NULL, NULL, $4, $5)`

	plannedAmount := category.PlannedAmount
	description := category.Description

	// Obter o status_id para "pending"
	var statusID string
	err = conn.QueryRow("SELECT status_id FROM status WHERE status_name = 'pending'").Scan(&statusID)
	if err != nil {
		// Rollback da inserção na tabela categories
		sqlQuery = `DELETE FROM categories WHERE category_id = $1`
		_, _ = conn.Exec(sqlQuery, category.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending status", "details": err.Error()})
		return
	}

	_, err = conn.Exec(monthlyExpenseQuery, category.ID, referenceMonth, plannedAmount, description, statusID)
	if err != nil {
		// Rollback da inserção na tabela categories
		sqlQuery = `DELETE FROM categories WHERE category_id = $1`
		_, _ = conn.Exec(sqlQuery, category.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert into monthly_expenses", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "Successfully created category and added to monthly expenses",
		"categoryId": category.ID,
	})
}

func DeleteCategory(ctx *gin.Context) {
	categoryID := ctx.Param("id")

	conn, err := db.OpenConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error connecting to database",
			"error":   err.Error(),
		})
		return
	}
	defer conn.Close()

	sqlQuery := `DELETE FROM categories WHERE category_id = $1`

	_, err = conn.Exec(sqlQuery, categoryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data from database"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Successfully deleted category",
	})
}

func UpdateCategory(ctx *gin.Context) {
	categoryID := ctx.Param("id") // Obtém o ID da categoria a ser atualizada

	var category Category

	// Fazer o bind dos dados recebidos no JSON para a struct `Category`
	if err := ctx.ShouldBindJSON(&category); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	conn, err := db.OpenConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error connecting to database",
			"error":   err.Error(),
		})
		return
	}
	defer conn.Close()

	// Verificar se a categoria existe antes de tentar atualizar
	var existingCategoryID string
	err = conn.QueryRow("SELECT category_id FROM categories WHERE category_id = $1", categoryID).Scan(&existingCategoryID)
	if err == sql.ErrNoRows {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Category not found"})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Error checking category existence", "details": err.Error()})
		return
	}

	// Atualizar a tabela `categories`
	sqlUpdateCategory := `UPDATE categories 
                          SET category_name = $1, amount_planned = $2, category_color = $3, description = $4 
                          WHERE category_id = $5`
	_, err = conn.Exec(sqlUpdateCategory, category.Name, category.PlannedAmount, category.Color, category.Description, categoryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update category", "details": err.Error()})
		return
	}

	// Atualizar a tabela `monthly_expenses` para o mês atual
	currentMonth := time.Now().Format("2006-01-01") // Formato YYYY-MM-01 para o PostgreSQL

	sqlUpdateExpense := `UPDATE monthly_expenses 
                         SET amount_planned = $1, description = $2 
                         WHERE category_id = $3 AND reference_month = $4`
	_, err = conn.Exec(sqlUpdateExpense, category.PlannedAmount, category.Description, categoryID, currentMonth)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update monthly expense", "details": err.Error()})
		return
	}

	// Retornar resposta de sucesso
	ctx.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Successfully updated category and monthly expense",
		"categoryId": categoryID,
	})
}

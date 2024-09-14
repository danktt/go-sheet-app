package analytic

import (
	"database/sql"
	"go-sheet/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// get planned, spent and diferenc amount by month
func GetAnalyticTotal(ctx *gin.Context) {
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

	// Obter o mês da query string, se fornecido
	monthParam := ctx.DefaultQuery("month", "")
	var targetMonth time.Time

	if monthParam != "" {
		// Se um mês foi fornecido, parse-o
		targetMonth, err = time.Parse("2006-01", monthParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid month format. Use YYYY-MM",
				"error":   err.Error(),
			})
			return
		}
	} else {
		// Se nenhum mês foi fornecido, use o mês atual
		targetMonth = time.Now()
	}

	// Formatar o mês alvo para o primeiro dia do mês
	startOfMonth := targetMonth.Format("2006-01-01")
	// Calcular o primeiro dia do próximo mês
	endOfMonth := targetMonth.AddDate(0, 1, 0).Format("2006-01-01")

	sqlQuery := `
		SELECT 
			SUM(amount_planned) AS total_planned, 
			SUM(COALESCE(spent_amount, 0)) AS total_spent, 
			SUM(amount_planned - COALESCE(spent_amount, 0)) AS total_difference 
		FROM monthly_expenses 
		WHERE reference_month >= $1 AND reference_month < $2
	`

	var totalPlanned, totalSpent, totalDifference sql.NullFloat64

	err = conn.QueryRow(sqlQuery, startOfMonth, endOfMonth).Scan(&totalPlanned, &totalSpent, &totalDifference)
	if err != nil {
		if err == sql.ErrNoRows {
			// Não há dados para o mês especificado, retornar zeros
			ctx.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "No data for the specified month",
				"data": gin.H{
					"month":           targetMonth.Format("2006-01"),
					"totalPlanned":    0,
					"totalSpent":      0,
					"totalDifference": 0,
				},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error executing query",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Analytic data retrieved successfully",
		"data": gin.H{
			"month":           targetMonth.Format("2006-01"),
			"totalPlanned":    totalPlanned.Float64,
			"totalSpent":      totalSpent.Float64,
			"totalDifference": totalDifference.Float64,
		},
	})
}

// GetPendingPayment retrieves all expenses with the status "pending"
// GetPendingPayment retrieves all pending payments for a specific month
func GetPendingPayment(ctx *gin.Context) {
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

	// Obter o mês da query string, se fornecido
	monthParam := ctx.DefaultQuery("month", "")
	var targetMonth time.Time

	if monthParam != "" {
		// Se um mês foi fornecido, parse-o
		targetMonth, err = time.Parse("2006-01", monthParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid month format. Use YYYY-MM",
				"error":   err.Error(),
			})
			return
		}
	} else {
		// Se nenhum mês foi fornecido, use o mês atual
		targetMonth = time.Now()
	}

	// Formatar o mês alvo para o primeiro dia do mês
	startOfMonth := targetMonth.Format("2006-01-01")
	// Calcular o primeiro dia do próximo mês
	endOfMonth := targetMonth.AddDate(0, 1, 0).Format("2006-01-01")

	// Query para buscar todas as despesas com o status "pending" e para o mês especificado
	sqlQuery := `
        SELECT 
            me.expense_id, 
            me.category_id, 
            c.category_name,
            me.reference_month, 
            me.spent_amount, 
            me.amount_planned, 
            me.payment_date, 
            me.description,
            s.status_name
        FROM 
            monthly_expenses me
        JOIN
            categories c ON me.category_id::text = c.category_id::text
        JOIN
            status s ON me.status_id::text = s.status_id::text
        WHERE 
            s.status_name = 'pending' AND me.reference_month >= $1 AND me.reference_month < $2
    `

	rows, err := conn.Query(sqlQuery, startOfMonth, endOfMonth)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error querying database",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	var pendingPayments []gin.H
	for rows.Next() {
		var expenseID, categoryID, categoryName, statusName string
		var description sql.NullString // Alterado para sql.NullString
		var referenceMonth time.Time
		var spentAmount, amountPlanned sql.NullFloat64
		var paymentDate sql.NullTime

		err := rows.Scan(&expenseID, &categoryID, &categoryName, &referenceMonth, &spentAmount, &amountPlanned, &paymentDate, &description, &statusName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Error scanning row",
				"error":   err.Error(),
			})
			return
		}

		payment := gin.H{
			"expenseId":      expenseID,
			"categoryId":     categoryID,
			"categoryName":   categoryName,
			"referenceMonth": referenceMonth.Format("2006-01-02"),
			"spentAmount":    spentAmount.Float64,
			"plannedAmount":  amountPlanned.Float64,
			"paymentDate":    paymentDate.Time.Format("2006-01-02"),
			"description":    description.String, // Use description.String
			"statusName":     statusName,
		}

		// Adicione a descrição apenas se ela não for nula
		if description.Valid {
			payment["description"] = description.String
		} else {
			payment["description"] = "" // ou qualquer valor padrão que você queira usar
		}

		pendingPayments = append(pendingPayments, payment)
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error iterating over rows",
			"error":   err.Error(),
		})
		return
	}

	if len(pendingPayments) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "No pending payments found for the specified month",
			"data":    []gin.H{},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Pending payments retrieved successfully",
		"data":    pendingPayments,
	})
}

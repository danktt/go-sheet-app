package handlers

import (
	"database/sql"
	"go-sheet/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MonthlyExpense struct {
	CategoryID     string  `json:"categoryId" binding:"required"`
	ReferenceMonth string  `json:"referenceMonth" binding:"required"`
	PaidId         string  `json:"paidId" binding:"required"`
	SpentAmount    float64 `json:"spentAmount" binding:"required"`
	PaymentDate    string  `json:"paymentDate" binding:"required"`
	File           string  `json:"file"`
}

// Update the MonthlyExpenseResponse struct
type MonthlyExpenseResponse struct {
	ExpenseID      string   `json:"expenseId"`
	CategoryName   string   `json:"categoryName"`
	ReferenceMonth *string  `json:"referenceMonth"` // colocar * significa que o campo é opcional
	SpentAmount    *float64 `json:"spentAmount"`
	PlannedAmount  float64  `json:"plannedAmount"`
	Difference     *float64 `json:"difference"`
	PaymentDate    *string  `json:"paymentDate"`
	File           *string  `json:"file"`
	PaidId         *string  `json:"paidId"`
	PaidType       *string  `json:"paidType"`  // Add this field
	PaidColor      *string  `json:"paidColor"` // Change this to a pointer
	StatusId       *string  `json:"statusId"`
	StatusName     *string  `json:"statusName"`
	Description    *string  `json:"description"`
}

// ListMonthlyExpenses retrieves all monthly expenses with category details
func ListMonthlyExpenses(ctx *gin.Context) {

	conn, err := db.OpenConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer conn.Close()

	// Update the query in the ListMonthlyExpenses function
	query := `
		SELECT 
			me.expense_id,
			c.category_name,
			me.reference_month,
			me.spent_amount,
			me.amount_planned,
			(me.amount_planned - me.spent_amount) AS difference,
			me.payment_date,
			me.file,
			pt.paid_id,
			pt.paid_type AS paid_type,
			pt.paid_color AS paid_color,
			st.status_id AS status_id,
			st.status_name AS status_name,
			me.description AS description
		FROM 
			monthly_expenses me
		JOIN 
			categories c ON me.category_id = c.category_id
		LEFT JOIN
			paid_type pt ON me.paid_id::text = pt.paid_id::text
		LEFT JOIN
			status st ON me.status_id::text = st.status_id::text
		ORDER BY 
			me.reference_month DESC;
	`

	rows, err := conn.Query(query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query expenses", "details": err.Error()})
		return
	}
	defer rows.Close()

	var expenses []MonthlyExpenseResponse

	for rows.Next() {
		var expense MonthlyExpenseResponse
		var spentAmount, difference sql.NullFloat64
		var paymentDate, file, paidId, paidType, paidColor sql.NullString
		var referenceMonth sql.NullTime // Mudança aqui
		var statusId, statusName, description sql.NullString
		err := rows.Scan(
			&expense.ExpenseID,
			&expense.CategoryName,
			&referenceMonth,
			&spentAmount,
			&expense.PlannedAmount,
			&difference,
			&paymentDate,
			&file,
			&paidId,
			&paidType,
			&paidColor,
			&statusId,
			&statusName,
			&description,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row", "details": err.Error()})
			return
		}

		if spentAmount.Valid {
			expense.SpentAmount = &spentAmount.Float64
		}
		if difference.Valid {
			expense.Difference = &difference.Float64
		}
		if paymentDate.Valid {
			expense.PaymentDate = &paymentDate.String
		}
		if file.Valid {
			expense.File = &file.String
		}
		if paidId.Valid {
			expense.PaidId = &paidId.String
		}
		if paidType.Valid {
			expense.PaidType = &paidType.String
		}
		if paidColor.Valid {
			expense.PaidColor = &paidColor.String
		}
		if statusId.Valid {
			expense.StatusId = &statusId.String
		}
		if statusName.Valid {
			expense.StatusName = &statusName.String
		}
		if description.Valid {
			expense.Description = &description.String
		}

		if referenceMonth.Valid {
			formattedDate := referenceMonth.Time.Format("2006-01-02")
			expense.ReferenceMonth = &formattedDate
		}

		expenses = append(expenses, expense)
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to iterate over rows", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Expenses retrieved successfully",
		"expenses": expenses,
	})
}

// CreateExpense inserts a new monthly expense into the database
func CreateExpense(ctx *gin.Context) {
	var expense MonthlyExpense

	// Bind the JSON received to the MonthlyExpense struct
	if err := ctx.ShouldBindJSON(&expense); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}

	conn, err := db.OpenConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database", "status": "error"})
		return
	}
	defer conn.Close()

	// Validate and convert the dates
	refMonth, err := time.Parse("2006-01-02", expense.ReferenceMonth)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reference month format. Use YYYY-MM-DD", "status": "error"})
		return
	}

	payDate, err := time.Parse("2006-01-02", expense.PaymentDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment date format. Use YYYY-MM-DD", "status": "error"})
		return
	}

	// Fetch amount_planned from the categories table to insert it into the monthly_expenses
	var amountPlanned float64
	err = conn.QueryRow(`SELECT amount_planned FROM categories WHERE category_id = $1`, expense.CategoryID).Scan(&amountPlanned)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve planned amount for the category", "status": "error"})
		return
	}

	// Generate a UUID for the new expense
	newUUID := uuid.New()

	// Insert into the monthly_expenses table
	sqlQuery := `INSERT INTO monthly_expenses (expense_id, category_id, reference_month, spent_amount, amount_planned, payment_date, paid_id, file) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = conn.Exec(sqlQuery, newUUID, expense.CategoryID, refMonth, expense.SpentAmount, amountPlanned, payDate, expense.PaidId, expense.File)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert expense", "status": "error"})
		return
	}

	// Return success with the generated UUID
	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Expense created successfully",
		"expense_id": newUUID,
		"status":     "success",
	})
}

package paid_type

import (
	"go-sheet/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaidType struct {
	ID        string `json:"uuid"`
	Type      string `json:"type"`
	PaidColor string `json:"color"`
	CreatedAt string `json:"createdAt"`
}

func ListPaidTypes(ctx *gin.Context) {
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

	rows, err := conn.Query("SELECT * FROM paid_type")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error querying database",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	var paidTypes []PaidType
	for rows.Next() {
		var paidType PaidType
		err := rows.Scan(&paidType.ID, &paidType.Type, &paidType.PaidColor, &paidType.CreatedAt)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Error scanning database",
				"error":   err.Error(),
			})
			return
		}
		paidTypes = append(paidTypes, paidType)
	}
	if len(paidTypes) == 0 {
		paidTypes = []PaidType{}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Paid types fetched successfully",
		"data":    paidTypes,
	})
}

func CreatePaidType(ctx *gin.Context) {
	var paidType PaidType

	if err := ctx.ShouldBindJSON(&paidType); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
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

	query := "INSERT INTO paid_type (paid_type, paid_color) VALUES ($1, $2) RETURNING paid_id"
	err = conn.QueryRow(query, paidType.Type, paidType.PaidColor).Scan(&paidType.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error creating paid type",
			"error":   err.Error(),
		})
		return
	}

	// Wrap the newly created paidType in an array
	paidTypes := []PaidType{paidType}

	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Paid type created successfully",
		"data":    paidTypes,
	})
}

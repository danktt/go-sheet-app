package status

import (
	"database/sql"
	"go-sheet/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Status struct {
	ID         string `json:"uuid"`
	StatusName string `json:"statusName"`
}

func ListStatus(ctx *gin.Context) {
	conn, err := db.OpenConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error connecting to database",
			"error":   err.Error(),
		})
	}
	defer conn.Close()

	sqlQuery := `SELECT * FROM status`
	rows, err := conn.Query(sqlQuery)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error querying database",
			"error":   err.Error(),
		})
	}

	var statuses []Status
	for rows.Next() {
		var status Status
		err = rows.Scan(&status.ID, &status.StatusName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Error scanning database rows",
				"error":   err.Error(),
			})
		}
		statuses = append(statuses, status)
	}

	rows.Close()

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Status list retrieved successfully",
		"data":    statuses,
	})
}

func CreateStatus(ctx *gin.Context) {
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

	var status Status
	err = ctx.ShouldBindJSON(&status)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Error binding JSON",
			"error":   err.Error(),
		})
		return
	}

	// Check if status with the same name already exists
	var existingID string
	checkQuery := `SELECT status_id FROM status WHERE status_name = $1`
	err = conn.QueryRow(checkQuery, status.StatusName).Scan(&existingID)
	if err == nil {
		ctx.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": "Status with this name already exists",
		})
		return
	} else if err != sql.ErrNoRows {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error checking existing status",
			"error":   err.Error(),
		})
		return
	}

	// If no existing status found, proceed with insertion
	sqlQuery := `INSERT INTO status (status_name) VALUES ($1) RETURNING status_id`
	err = conn.QueryRow(sqlQuery, status.StatusName).Scan(&status.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error inserting data into database",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Status created successfully",
		"data":    status,
	})
}

func DeleteStatus(ctx *gin.Context) {
	conn, err := db.OpenConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error connecting to database",
			"error":   err.Error(),
		})
	}
	defer conn.Close()

	statusID := ctx.Param("id")

	sqlQuery := `DELETE FROM status WHERE status_id = $1`
	_, err = conn.Exec(sqlQuery, statusID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error deleting status",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Status deleted successfully",
	})
}

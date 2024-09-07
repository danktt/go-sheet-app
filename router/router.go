package routes

import (
	"github.com/gin-gonic/gin"
)

func Initialize(r *gin.Engine) {

	server := gin.Default()

	InitializeRoutes(server)

	server.Run(":8080")
}

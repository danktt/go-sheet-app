package main

import (
	routes "go-sheet/router"

	"github.com/gin-gonic/gin"
)

func main() {

	routes.Initialize(gin.Default())

}

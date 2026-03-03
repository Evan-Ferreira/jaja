package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHelloWorld(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, "It's JAJA Time")
}

func main() {
    router := gin.Default()
	
    router.GET("/", getHelloWorld)

    router.Run("localhost:8080")
}
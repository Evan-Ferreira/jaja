package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type D2LHandler struct {
}

func (h *D2LHandler) Hello(c *gin.Context) {
	fmt.Println("Hello, D2L!")
	c.String(http.StatusOK, "Hello, D2L!")
}

package main

import (
	"backend/internal/api"
	"backend/internal/api/handlers"
	"fmt"
)

func main() {
	d2lHandler := &handlers.D2LHandler{}
	router := api.NewRouter(d2lHandler)
	fmt.Println("Starting server on http://localhost:8080")

	router.Run("localhost:8080")
}

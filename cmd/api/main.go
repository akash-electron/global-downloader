package main

import (
	"fmt"
	"net/http"

	"global-downloader/internal/routes"
)

func main() {

	router := routes.SetupRoutes()

	fmt.Println("Server running on :8080")

	http.ListenAndServe(":8080", router)
}
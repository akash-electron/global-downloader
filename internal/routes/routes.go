 package routes

import (
	"net/http"

	"global-downloader/internal/handlers"
)

func SetupRoutes() *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("/download", handlers.Download)

	return mux
}
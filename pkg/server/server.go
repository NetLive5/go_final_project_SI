package server

import (
	"fmt"
	"net/http"

	"go_final_project/pkg/api"
)


func Run(port, webDir string) error {
	if port == "" {
		return fmt.Errorf("port must not be empty")
	}

	mux := http.NewServeMux()
	api.Init(mux)
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	return http.ListenAndServe(":"+port, mux)
}

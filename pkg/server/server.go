package server

import (
	"fmt"
	"net/http"
)

// Run serves the frontend files from webDir on the requested TCP port.
func Run(port, webDir string) error {
	if port == "" {
		return fmt.Errorf("port must not be empty")
	}

	handler := http.FileServer(http.Dir(webDir))
	return http.ListenAndServe(":"+port, handler)
}

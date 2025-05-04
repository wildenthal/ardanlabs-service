package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type httpController struct {
	logger *slog.Logger
}

func NewHTTPController(logger *slog.Logger) *httpController {
	return &httpController{
		logger: logger,
	}
}

func (c *httpController) StatusOKHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(map[string]string{"Status": "OK"})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	c.logger.InfoContext(r.Context(), "Looking good", "method", r.Method, "path", r.URL.Path)
}

// PanicHandler is used to intentionally trigger a panic for testing middleware recovery mechanisms.
func (c *httpController) PanicHandler(w http.ResponseWriter, r *http.Request) {
	panic("This is a panic")
}

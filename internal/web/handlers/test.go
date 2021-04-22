package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/strax84mb/go-travel-reactive/internal/web"
)

func RegisterTestHandler(r *mux.Router) {
	r.Methods(http.MethodGet).Path("/test").HandlerFunc(testHandler())
}

func testHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		web.Ok(w, struct {
			Status string `json:"status"`
		}{
			Status: "ok",
		})
	}
}

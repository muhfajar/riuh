package handler

import (
	"github.com/muhfajar/riuh/api"
	"net/http"
)

func H(w http.ResponseWriter, r *http.Request) {
	api.Routes().ServeHTTP(w, r)
}

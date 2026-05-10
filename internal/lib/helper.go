package lib

import (
	"encoding/json"
	"net/http"

	"github.com/Bughay/go-backend-layout/internal/model"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, model.APIError{Code: status, Message: message})
}

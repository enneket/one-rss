package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func HandleError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*AppError); ok {
		respondWithError(w, appErr.Code, appErr.Message)
		return
	}

	log.Printf("Internal error: %v", err)
	respondWithError(w, http.StatusInternalServerError, "Internal server error")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func RespondSuccess(w http.ResponseWriter, message string) {
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": message})
}

func RespondBadRequest(w http.ResponseWriter, message string) {
	respondWithError(w, http.StatusBadRequest, message)
}

func RespondNotFound(w http.ResponseWriter, message string) {
	respondWithError(w, http.StatusNotFound, message)
}

func RespondInternalError(w http.ResponseWriter, message string) {
	respondWithError(w, http.StatusInternalServerError, message)
}

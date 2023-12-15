package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"wb-l0/pkg/httperrors"
)

type HandlerFunc = func(r *http.Request) (any, error)

func ErrorHandler(err error) http.HandlerFunc {
	return WrapHandler(func(r *http.Request) (any, error) {
		return nil, err
	})
}

func WrapHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := handler(r)
		if err != nil {
			sendError(w, err)
			return
		}

		status := http.StatusOK
		if data == nil {
			status = http.StatusNoContent
		}

		sendData(w, data, status)
	}
}

type errorResponse struct {
	Message string `json:"message"`
}

func sendError(w http.ResponseWriter, err error) {
	log.Println("error while handling request:", err)
	var httpError httperrors.Error

	if errors.As(err, &httpError) {
		sendData(w, errorResponse{httpError.Error()}, httpError.GetStatusCode())
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func sendData(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	err := encoder.Encode(data)
	if err != nil {
		log.Println("error writing response:", err)
		return
	}
}

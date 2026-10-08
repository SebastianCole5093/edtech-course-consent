package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	consent "example.com/edtech-consent-service"
)

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	client := consent.NewClient(apiKey)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /course-consent", courseConsentHandler(client))

	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = ":8080"
	}
	server := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("course consent service listening on %s", address)
	log.Fatal(server.ListenAndServe())
}

func courseConsentHandler(client *consent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var input consent.CourseRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if input.UserID == "" || input.CourseID == "" || input.RequestID == "" || input.Deadline.IsZero() {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_id, course_id, deadline, and request_id are required"})
			return
		}
		if input.Category == "" {
			input.Category = consent.LearningRecordsCategory
		}

		var err error
		switch strings.ToLower(input.Action) {
		case "grant":
			err = client.Grant(r.Context(), input.UserID, input.Category, input.RequestID)
			input.ConsentGranted = err == nil
		case "revoke":
			err = client.Revoke(r.Context(), input.UserID, input.Category, input.RequestID)
			input.ConsentGranted = false
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be grant or revoke"})
			return
		}
		if err != nil {
			var apiErr *consent.InfraiError
			if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
				writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Code, "message": apiErr.Message})
				return
			}
			log.Printf("consent operation failed: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "consent service request failed"})
			return
		}
		writeJSON(w, http.StatusOK, consent.DecideCourseAccess(input))
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

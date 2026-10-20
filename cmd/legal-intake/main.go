package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"legal-intake-inbox/intake"
)

func main() {
	teamInbox := os.Getenv("TEAM_INBOX")
	if teamInbox == "" {
		log.Fatal("TEAM_INBOX is required")
	}
	client, err := intake.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("POST /contact", func(w http.ResponseWriter, r *http.Request) {
		var form intake.MatterSubmission
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			respondError(w, http.StatusBadRequest, "invalid contact form")
			return
		}
		form.IP = r.RemoteAddr
		if form.RequestID == "" {
			respondError(w, http.StatusBadRequest, "request_id is required")
			return
		}
		result, err := client.RouteMatter(r.Context(), teamInbox, form)
		if err != nil {
			var apiErr *intake.APIError
			if errors.As(err, &apiErr) && apiErr.Status >= 400 && apiErr.Status < 500 {
				respondError(w, http.StatusUnprocessableEntity, "contact form rejected")
				return
			}
			respondError(w, http.StatusBadGateway, "unable to route contact form")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
	log.Println("legal intake listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

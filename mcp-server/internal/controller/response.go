package controller

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondWithError sends an error response
func RespondWithError(w http.ResponseWriter, code int, message string, err error) {
	log.Error().Err(err).Str("message", message).Msg("Request error")
	
	response := map[string]interface{}{
		"error":   message,
		"code":    code,
		"details": "",
	}
	
	if err != nil {
		response["details"] = err.Error()
	}

	RespondWithJSON(w, code, response)
}


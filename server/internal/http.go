package internal

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, statusCode int, model interface{}) {
	if json, err := json.Marshal(model); err != nil {
		log.Printf("Something went wrong returning a JSON response to the user %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{
			"status": 500,
			"message": "Something went wrong"
		}`))
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write(json)
	}
}

func WriteJSONError(w http.ResponseWriter, statusCode int, errorMessage string) {
	error := map[string]interface{}{
		"status":  statusCode,
		"message": errorMessage,
	}

	WriteJSON(w, statusCode, error)
}

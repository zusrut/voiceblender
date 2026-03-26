package api

import (
	"encoding/json"
	"net/http"
)

// instanceID is set by NewServer and injected into every JSON response body.
var instanceID string

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Inject instance_id as the first field in JSON object responses.
	if instanceID != "" && len(data) > 0 && data[0] == '{' {
		data = append([]byte(`{"instance_id":"`+instanceID+`",`), data[1:]...)
	}
	w.WriteHeader(status)
	w.Write(data)
	w.Write([]byte("\n"))
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

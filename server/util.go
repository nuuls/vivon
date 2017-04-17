package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	bs, err := json.Marshal(data)
	if err != nil {
		log.WithError(err).Error("")
		http.Error(w, http.StatusText(500), 500)
		return
	}
	w.Write(bs)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)

	writeJSON(w, struct {
		StatusCode int    `json:"statusCode"`
		Status     string `json:"status"`
		Message    string `json:"message,omitempty"`
	}{
		status,
		http.StatusText(status),
		msg,
	})
}

func unmarshalJSON(r io.Reader, v interface{}) error {
	if closer, ok := r.(io.Closer); ok {
		defer closer.Close()
	}
	return json.NewDecoder(r).Decode(v)
}

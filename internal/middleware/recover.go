package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bufferedWriter := newBufferedResponseWriter()

		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered: %v", recovered)
				writeError(w, http.StatusInternalServerError, "Internal server error")
				return
			}

			bufferedWriter.writeTo(w)
		}()

		next.ServeHTTP(bufferedWriter, r)
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorResponse{Error: message, Code: status}); err != nil {
		log.Printf("write error response: %v", err)
	}
}

type bufferedResponseWriter struct {
	header      http.Header
	body        bytes.Buffer
	statusCode  int
	wroteHeader bool
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{header: make(http.Header)}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.statusCode = statusCode
	w.wroteHeader = true
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.body.Write(data)
}

func (w *bufferedResponseWriter) writeTo(destination http.ResponseWriter) {
	for key, values := range w.header {
		destination.Header()[key] = append([]string(nil), values...)
	}

	statusCode := w.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	destination.WriteHeader(statusCode)
	if _, err := destination.Write(w.body.Bytes()); err != nil {
		log.Printf("write buffered response: %v", err)
	}
}

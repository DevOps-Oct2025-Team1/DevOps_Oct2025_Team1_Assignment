package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSHeaders(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		expectedOrigin  string
		expectedMethods string
		expectedHeaders string
	}{
		{
			name:            "GET request CORS headers",
			method:          http.MethodGet,
			expectedOrigin:  "*",
			expectedMethods: "GET, POST, PUT, DELETE, OPTIONS",
			expectedHeaders: "Content-Type, Authorization",
		},
		{
			name:            "POST request CORS headers",
			method:          http.MethodPost,
			expectedOrigin:  "*",
			expectedMethods: "GET, POST, PUT, DELETE, OPTIONS",
			expectedHeaders: "Content-Type, Authorization",
		},
		{
			name:            "PUT request CORS headers",
			method:          http.MethodPut,
			expectedOrigin:  "*",
			expectedMethods: "GET, POST, PUT, DELETE, OPTIONS",
			expectedHeaders: "Content-Type, Authorization",
		},
		{
			name:            "DELETE request CORS headers",
			method:          http.MethodDelete,
			expectedOrigin:  "*",
			expectedMethods: "GET, POST, PUT, DELETE, OPTIONS",
			expectedHeaders: "Content-Type, Authorization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			corsHandler := corsMiddleware(handler)

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			corsHandler.ServeHTTP(w, req)

			if got := w.Header().Get("Access-Control-Allow-Origin"); got != tt.expectedOrigin {
				t.Errorf("Access-Control-Allow-Origin = %v, want %v", got, tt.expectedOrigin)
			}
			if got := w.Header().Get("Access-Control-Allow-Methods"); got != tt.expectedMethods {
				t.Errorf("Access-Control-Allow-Methods = %v, want %v", got, tt.expectedMethods)
			}
			if got := w.Header().Get("Access-Control-Allow-Headers"); got != tt.expectedHeaders {
				t.Errorf("Access-Control-Allow-Headers = %v, want %v", got, tt.expectedHeaders)
			}
		})
	}
}
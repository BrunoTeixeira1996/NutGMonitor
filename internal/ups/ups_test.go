package ups

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
)

func TestValidateNutUPSContainer(t *testing.T) {
	// Setup logger
	err := logger.Setup("logs")
	if err != nil {
		t.Fatal("failed to setup logger:", err)
	}

	tests := []struct {
		name          string
		mockResponse  string
		expectedError string
	}{
		{
			name:          "VALID RESPONSE",
			mockResponse:  "OK",
			expectedError: "",
		},
		{
			name:          "EMPTY RESPONSE",
			mockResponse:  "",
			expectedError: "[ups info] nutupsd returned an empty body",
		},
		{
			name:          "BROKEN CONTAINER",
			mockResponse:  "Failed to connect to target",
			expectedError: "[ups info] nutupsd url might be broken: Failed to connect to target",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server with a mock response
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tt.mockResponse))
			}))
			defer ts.Close()

			// Call ValidateNutUPSContainer and check for error
			err := ValidateNutUPSContainer(ts.URL)

			// Check if the error message matches the expected one
			if err != nil && err.Error() != tt.expectedError {
				t.Errorf("expected error [%s] but got [%s]", tt.expectedError, err.Error())
			}
		})
	}
}

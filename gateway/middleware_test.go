package gateway

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-kit/log"
)

func TestAuthenticate(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stdout)
	testCases := []struct {
		name           string
		config         *Configuration
		authHeader     string
		expectedStatus int
	}{
		{
			name: "missing auth header",
			config: &Configuration{
				AuthType: "basic",
				Logger:   logger,
			},
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "missing basic auth credentials",
			config: &Configuration{
				AuthType: "basic",
				Logger:   logger,
			},
			authHeader:     "Bearer token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid credentials",
			config: &Configuration{
				AuthType: "basic",
				Tenants: []Tenant{
					{
						Username:    "username1",
						Password:    "password1",
						XScopeOrgId: "orgid",
					},
				},
				Routes:  []Route{},
				Targets: map[string]string{},
			},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("username1:password1")),
			expectedStatus: http.StatusOK,
		},
		{
			name: "wrong password",
			config: &Configuration{
				AuthType: "basic",
				Tenants: []Tenant{
					{
						Username:    "username1",
						Password:    "password1",
						XScopeOrgId: "orgid",
					},
				},
				Routes:  []Route{},
				Targets: map[string]string{},
				Logger:  logger,
			},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("username1:wrong_password")),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "no valid credentials",
			config: &Configuration{
				AuthType: "basic",
				Tenants: []Tenant{
					{
						Username:    "username1",
						Password:    "password1",
						XScopeOrgId: "orgid1",
					},
				},
				Routes:  []Route{},
				Targets: map[string]string{},
				Logger:  logger,
			},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("username2:password2")),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://localhost", nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			req.Header.Set("Authorization", tc.authHeader)

			rw := httptest.NewRecorder()

			tc.config.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(rw, req)

			if rw.Code != tc.expectedStatus {
				t.Errorf("expected status code %d, but got %d", tc.expectedStatus, rw.Code)
			}
		})
	}
}

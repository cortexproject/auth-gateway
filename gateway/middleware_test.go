package gateway

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	InitLogger(os.Stdout)
	testCases := []struct {
		name           string
		config         *Config
		authHeader     string
		expectedStatus int
	}{
		{
			name: "missing auth header",
			config: &Config{
				Tenants: []Tenant{
					{
						Authentication: "basic",
						Username:       "username1",
						Password:       "password1",
						ID:             "orgid1",
					},
				},
			},
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "missing basic auth credentials",
			config: &Config{
				Tenants: []Tenant{
					{
						Authentication: "basic",
						Username:       "username1",
						Password:       "password1",
						ID:             "orgid1",
					},
				},
			},
			authHeader:     "Bearer token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid credentials",
			config: &Config{
				Tenants: []Tenant{
					{
						Authentication: "basic",
						Username:       "username1",
						Password:       "password1",
						ID:             "orgid",
					},
				}},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("username1:password1")),
			expectedStatus: http.StatusOK,
		},
		{
			name: "wrong password",
			config: &Config{
				Tenants: []Tenant{
					{
						Authentication: "basic",
						Username:       "username1",
						Password:       "password1",
						ID:             "orgid",
					},
				},
			},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("username1:wrong_password")),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "no valid credentials",
			config: &Config{
				Tenants: []Tenant{
					{
						Authentication: "basic",
						Username:       "username1",
						Password:       "password1",
						ID:             "orgid1",
					},
				},
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

			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

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

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
		tenants        *Tenant
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "missing auth header",
			tenants:        &Tenant{},
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing basic auth credentials",
			tenants:        &Tenant{},
			authHeader:     "Bearer token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid credentials",
			tenants: &Tenant{
				All: map[string]tenant{
					"username1": {
						ID:       "orgid",
						Username: "username1",
						Password: "password1",
					},
				},
			},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("username1:password1")),
			expectedStatus: http.StatusOK,
		},
		{
			name: "wrong password",
			tenants: &Tenant{
				All: map[string]tenant{
					"username1": {
						ID:       "orgid",
						Username: "username1",
						Password: "password1",
					},
				},
				Logger: logger,
			},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("username1:wrong_password")),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "no valid credentials",
			tenants: &Tenant{
				All: map[string]tenant{
					"username1": {
						ID:       "orgid",
						Username: "username1",
						Password: "password1",
					},
				},
				Logger: logger,
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

			tc.tenants.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(rw, req)

			if rw.Code != tc.expectedStatus {
				t.Errorf("expected status code %d, but got %d", tc.expectedStatus, rw.Code)
			}
		})
	}
}

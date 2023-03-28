package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProxy(t *testing.T) {
	testCases := []struct {
		name       string
		targetURL  string
		expectErr  bool
		expectHost string
	}{
		{
			name:       "invalid URL",
			targetURL:  "invalid url",
			expectErr:  true,
			expectHost: "",
		},
		{
			name:       "valid URL",
			targetURL:  "http://localhost:8080",
			expectErr:  false,
			expectHost: "localhost:8080",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProxy(tc.targetURL)
			if (err != nil) != tc.expectErr {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tc.expectErr {
				if p != nil {
					t.Error("expected nil proxy, but got non-nil")
				}
				return
			}
			if p.targetURL.Host != tc.expectHost {
				t.Errorf("expected target host to be '%s', but got '%s'", tc.expectHost, p.targetURL.Host)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	testCases := []struct {
		name           string
		hasAuthHeader  bool
		expectedHeader string
		expectedStatus int
	}{
		{
			name:           "header present",
			hasAuthHeader:  true,
			expectedHeader: "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "header not present",
			hasAuthHeader:  false,
			expectedHeader: "Authorization",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proxy, err := NewProxy("http://example.com")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			req, err := http.NewRequest("GET", "http://example.com", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.hasAuthHeader {
				req.Header.Set("Authorization", "Bearer 123")
			}
			rr := httptest.NewRecorder()

			proxy.Handler(rr, req)

			if _, ok := req.Header[tc.expectedHeader]; ok {
				t.Errorf("Unexpected Authorization header found: %s", req.Header.Get("Authorization"))
			}

			if rr.Code != tc.expectedStatus {
				t.Errorf("Unexpected status code: %v", rr.Code)
			}
		})
	}
}

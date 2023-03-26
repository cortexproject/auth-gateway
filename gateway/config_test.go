package gateway

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/go-kit/log"
)

func TestInit(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stdout)
	testCases := []struct {
		name       string
		filePath   string
		logger     log.Logger
		configFile Config
		wantErr    error
	}{
		{
			name:     "Valid input file",
			filePath: "testdata/valid.yaml",
			logger:   logger,
			configFile: Config{
				ServerAddress: "localhost:8080",
				Tenants: []Tenant{
					{
						Authentication: "basic",
						Username:       "user1",
						Password:       "pass1",
						ID:             "1",
					},
				},
				Distributor: struct {
					URL   string   `yaml:"url"`
					Paths []string `yaml:"paths"`
				}{
					URL: "http://localhost:8081",
					Paths: []string{
						"/api/v1",
						"/api/v1/push",
					},
				}},
			wantErr: nil,
		},
		{
			name:       "Invalid input file",
			filePath:   "testdata/invalid.yaml",
			logger:     logger,
			configFile: Config{},
			wantErr:    errors.New("line 8: cannot unmarshal"),
		},
		{
			name:       "Non-existent input file",
			filePath:   "testdata/nonexistent.yaml",
			logger:     logger,
			configFile: Config{},
			wantErr:    errors.New("no such file or directory"),
		},
		{
			name:       "Empty input file",
			filePath:   "testdata/empty.yaml",
			logger:     logger,
			configFile: Config{},
			wantErr:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf, err := Init(tc.filePath)
			if !reflect.DeepEqual(conf, tc.configFile) {
				t.Errorf("Unexpected result: got %v, want %v", conf, tc.configFile)
			}

			if err == nil && tc.wantErr != nil || err != nil && tc.wantErr == nil {
				t.Errorf("Unexpected error: got %v, want %v", err, tc.wantErr)
			}

			if err != nil && tc.wantErr != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				t.Errorf("Unexpected error: got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

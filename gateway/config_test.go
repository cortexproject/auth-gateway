package gateway

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/go-kit/log"
)

func TestInit(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stdout)
	testCases := []struct {
		name       string
		filePath   string
		logger     log.Logger
		configFile Configuration
		wantErr    error
	}{
		{
			name:     "Valid input file",
			filePath: "testdata/valid.yaml",
			logger:   logger,
			configFile: Configuration{
				ServerAddress: "localhost:8080",
				AuthType:      "basic",
				Routes: []Route{
					{
						Path:   "/api/v1",
						Target: "http://localhost:8081",
					},
				},
				Logger: nil,
			},
			wantErr: nil,
		},
		{
			name:     "Invalid input file",
			filePath: "testdata/invalid.yaml",
			logger:   logger,
			configFile: Configuration{
				Logger: nil,
			},
			wantErr: errors.New("yaml: line 4: did not find expected key"),
		},
		{
			name:     "Non-existent input file",
			filePath: "testdata/nonexistent.yaml",
			logger:   logger,
			configFile: Configuration{
				Logger: nil,
			},
			wantErr: &os.PathError{Op: "open", Path: "testdata/nonexistent.yaml", Err: errors.New("no such file or directory")},
		},
		{
			name:     "Empty input file",
			filePath: "testdata/empty.yaml",
			logger:   logger,
			configFile: Configuration{
				Logger: nil,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf, err := Init(tc.filePath, nil)
			if !reflect.DeepEqual(conf, tc.configFile) {
				t.Errorf("Unexpected result: got %v, want %v", conf, tc.configFile)
			}

			if err == nil && tc.wantErr != nil || err != nil && tc.wantErr == nil {
				t.Errorf("Unexpected error: got %v, want %v", err, tc.wantErr)
			}

			if err != nil && tc.wantErr != nil && err.Error() != tc.wantErr.Error() {
				t.Errorf("Unexpected error: got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

package gateway

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestGetTenants(t *testing.T) {
	testCases := []struct {
		name            string
		filePath        string
		expectedTenants Tenant
		expectedErr     error
	}{
		{
			name:     "Valid input file",
			filePath: "testdata/valid.yaml",
			expectedTenants: Tenant{
				All: map[string]tenant{
					"username1": {
						ID:       "orgid1",
						Username: "username1",
						Password: "password1",
					},
					"username2": {
						ID:       "orgid2",
						Username: "username2",
						Password: "password2",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:            "Invalid input file",
			filePath:        "testdata/invalid.yaml",
			expectedTenants: Tenant{},
			expectedErr:     errors.New("yaml: unmarshal errors:\n  line 2: cannot unmarshal !!seq into map[string]gateway.tenant"),
		},
		{
			name:            "Non-existent input file",
			filePath:        "testdata/nonexistent.yaml",
			expectedTenants: Tenant{},
			expectedErr:     &os.PathError{Op: "open", Path: "testdata/nonexistent.yaml", Err: errors.New("no such file or directory")},
		},
		{
			name:     "Empty input file",
			filePath: "testdata/empty.yaml",
			expectedTenants: Tenant{
				All: map[string]tenant{},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tenants, err := InitTenants(tc.filePath, nil)
			if !reflect.DeepEqual(tenants, tc.expectedTenants) {
				t.Errorf("Unexpected result: got %v, want %v", tenants, tc.expectedTenants)
			}
			if err == nil && tc.expectedErr != nil || err != nil && tc.expectedErr == nil {
				t.Errorf("Unexpected error: got %v, want %v", err, tc.expectedErr)
			}
			if err != nil && tc.expectedErr != nil && err.Error() != tc.expectedErr.Error() {
				t.Errorf("Unexpected error: got %v, want %v", err, tc.expectedErr)
			}
		})
	}
}

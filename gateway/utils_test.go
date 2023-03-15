package gateway

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadYaml(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "example.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	_, err = tmpfile.Write([]byte(`
basic:
  credential1:
    username: user1
    password: password1
    tenantID: tenant1
  credential2:
    username: user2
    password: password2
    tenantID: tenant2
`))
	if err != nil {
		t.Fatal(err)
	}

	// close the file before passing the file path to ReadYaml
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// test ReadYaml with the temporary file path
	config, err := ReadYaml(tmpfile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// test that the Basic field is correctly populated
	expectedCredential1 := Credentials{
		Username: "user1",
		Password: "password1",
		TenantID: "tenant1",
	}
	expectedCredential2 := Credentials{
		Username: "user2",
		Password: "password2",
		TenantID: "tenant2",
	}
	assert.Equal(t, expectedCredential1, config.Basic["credential1"])
	assert.Equal(t, expectedCredential2, config.Basic["credential2"])
}

package gateway

import (
	"bytes"
	"errors"
	"testing"

	"github.com/go-kit/log"
)

func TestCheckErr(t *testing.T) {
	var buf bytes.Buffer
	logger := log.NewJSONLogger(&buf)

	exitCalled := false
	mockExit := func(code int) {
		exitCalled = true
	}

	CheckErrWithExit("test message", nil, logger, mockExit)
	if exitCalled {
		t.Error("CheckErrWithExit called os.Exit for a nil error")
	}

	err := errors.New("test error")
	CheckErrWithExit("test message", err, logger, mockExit)
	if !exitCalled {
		t.Error("CheckErrWithExit did not call os.Exit for a non-nil error")
	}

	logOutput := buf.String()

	expectedMsgOutput := `"msg":"error test message"`
	expectedErrOutput := `"err":"test error"`

	if logOutput == "" || !(bytes.Contains([]byte(logOutput), []byte(expectedMsgOutput)) &&
		bytes.Contains([]byte(logOutput), []byte(expectedErrOutput))) {
		t.Errorf("Expected log to contain both '%s' and '%s', but got '%s'", expectedMsgOutput, expectedErrOutput, logOutput)
	}
}

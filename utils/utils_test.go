package utils

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type exitFuncMock struct {
	called bool
	code   int
}

func (e *exitFuncMock) exitFunc(code int) {
	e.called = true
	e.code = code
}

func TestLogrusErrorWriter_Write(t *testing.T) {
	testErr := []byte("test error message")

	buf := &bytes.Buffer{}
	logrus.SetOutput(buf)
	defer logrus.SetOutput(os.Stderr)

	writer := LogrusErrorWriter{}
	n, err := writer.Write(testErr)

	assert.NoError(t, err)
	assert.Equal(t, len(testErr), n)

	assert.True(t, strings.Contains(buf.String(), "test error message"))
}

func TestCheckErr(t *testing.T) {
	err := errors.New("test error")
	exitMock := &exitFuncMock{}

	buf := &bytes.Buffer{}
	logrus.SetOutput(buf)
	defer logrus.SetOutput(os.Stderr)

	CheckErrWithExit("An error occurred", err, exitMock.exitFunc)

	assert.True(t, strings.Contains(buf.String(), "test error"))
	assert.True(t, exitMock.called)
	assert.Equal(t, 1, exitMock.code)
}

func TestCheckErr_NoError(t *testing.T) {
	err := error(nil)
	exitMock := &exitFuncMock{}

	buf := &bytes.Buffer{}
	logrus.SetOutput(buf)
	defer logrus.SetOutput(os.Stderr)

	CheckErrWithExit("No error should occur", err, exitMock.exitFunc)

	assert.False(t, strings.Contains(buf.String(), "No error should occur"))
	assert.False(t, exitMock.called)
}

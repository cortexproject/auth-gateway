package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

type LogrusErrorWriter struct{}

func (w LogrusErrorWriter) Write(p []byte) (n int, err error) {
	logrus.Errorf("%s", string(p))
	return len(p), nil
}

func CheckErr(msg string, err error) {
	CheckErrWithExit(msg, err, os.Exit)
}

func CheckErrWithExit(msg string, err error, exitFunc func(int)) {
	if err != nil {
		logrus.Errorf("err: %v", err)
		exitFunc(1)
	}
}

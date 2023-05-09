package utils

import "github.com/sirupsen/logrus"

type LogrusErrorWriter struct{}

func (w LogrusErrorWriter) Write(p []byte) (n int, err error) {
	logrus.Errorf("%s", string(p))
	return len(p), nil
}

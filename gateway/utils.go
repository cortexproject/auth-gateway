package gateway

import (
	"fmt"
	"io"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var logger log.Logger

func InitLogger(output io.Writer) {
	logger = log.NewLogfmtLogger(output)
}

func CheckErr(msg string, err error) {
	CheckErrWithExit(msg, err, os.Exit)
}

func CheckErrWithExit(msg string, err error, exitFunc func(int)) {
	if err != nil {
		logger := log.With(level.Error(logger), "msg", "error "+msg)
		logger.Log("err", fmt.Sprintf("%+v", err))
		exitFunc(1)
	}
}

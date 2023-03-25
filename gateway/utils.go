package gateway

import (
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func CheckErr(msg string, err error, logger log.Logger) {
	CheckErrWithExit(msg, err, logger, os.Exit)
}

func CheckErrWithExit(msg string, err error, logger log.Logger, exitFunc func(int)) {
	if err != nil {
		logger := log.With(level.Error(logger), "msg", "error "+msg)
		logger.Log("err", fmt.Sprintf("%+v", err))
		exitFunc(1)
	}
}

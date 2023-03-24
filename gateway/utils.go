package gateway

import (
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func CheckErr(msg string, err error, logger log.Logger) {
	if err != nil {
		logger := level.Error(logger)
		if msg != "" {
			logger = log.With(logger, "msg", "error "+msg)
		}
		logger.Log("err", fmt.Sprintf("%+v", err))
		os.Exit(1)
	}
}

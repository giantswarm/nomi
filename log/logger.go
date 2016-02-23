package log

import (
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("fleemer")

func init() {
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	backendError := logging.NewLogBackend(os.Stderr, "", 0)
	backendLogs := logging.NewLogBackend(os.Stderr, "", 0)

	backendErrorLeveled := logging.AddModuleLevel(backendError)
	backendErrorLeveled.SetLevel(logging.ERROR, "")

	backendFormatter := logging.NewBackendFormatter(backendLogs, format)
	logging.SetBackend(backendErrorLeveled, backendFormatter)
}

func Logger() *logging.Logger {
	return log
}

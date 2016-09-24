package main

import (
    "github.com/op/go-logging"
    "os"
)

var log = logging.MustGetLogger("main")

var format = logging.MustStringFormatter(
    `%{color}%{time:15:04:05.000} %{shortfunc:16s} > %{level:.4s}:%{color:reset} %{message}`,
)

func SetupLogger(consoleLevel logging.Level) {
    consoleBackend := logging.NewLogBackend(os.Stderr, "", 0)

    consoleBackendFormatter := logging.NewBackendFormatter(consoleBackend, format)
    consoleBackendLeveled := logging.AddModuleLevel(consoleBackendFormatter)
    consoleBackendLeveled.SetLevel(consoleLevel, "")

    logging.SetBackend(consoleBackendLeveled)
}

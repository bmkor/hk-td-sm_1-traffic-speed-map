package main

import (
	"time"

	"github.com/yieldr/go-log/log"
	"github.com/yieldr/go-log/log/logrotate"
)

func createLogger() (*logrotate.Logrotate, error) {
	sink, err := logrotate.New("log/download.log", time.Hour, log.BasicFormat, log.BasicFields)
	if err != nil {
		panic(err)
	}
	Logger = log.New(sink)
	notify.Notify("Traffic Speed Data", "Log created")
	Logger.Info("Download log created.")
	go sink.Run()
	return sink, err
}

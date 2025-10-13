package main

import (
	"fmt"
	"log/slog"
	"os"
)

func createAppDataFolder(applicationName string) string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	dir = dir + "\\" + applicationName
	err = os.MkdirAll(dir, 0600)
	if err != nil {
		return ""
	}
	return dir
}

func openLogFile(fileName string) *os.File {
	fi, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// log file not ready so default std.err logging here
		slog.Error(fmt.Sprintf("%s\n", "Failed to open logfile for writing"))
		slog.Error(err.Error())
		return &os.File{}
	}
	return fi
}

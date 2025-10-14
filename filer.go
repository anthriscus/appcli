package main

import (
	"fmt"
	"log/slog"
	"os"
)

func createAppDataFolder(applicationName string) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir = dir + "\\" + applicationName
	err = os.MkdirAll(dir, 0600)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func openLogFile(fileName string) (*os.File, error) {
	fi, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// log file not ready so default std.err logging here
		slog.Error(fmt.Sprintf("%s\n", "Failed to create logfile for writing"))
		slog.Error(err.Error())
		return &os.File{}, err
	}
	return fi, nil
}

package main

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/log"
	"os"
)

var (
	logger *Logger
)

const (
	workingDirectory = "./segmentio/consumer/logs/"
)

// Logger is struct to log info not only to console, but in files too. InConsole function can be called right after LogF
// It works with last file (now it's 3, but can be more), files managing described in SetUpLogger
// IMPORTANT: Needs logs1.log, logs2.log and logs3.log to be created in logs folder
type Logger struct {
	file *os.File
}

type LastMessage string

func SetUpLogger() error {
	logger = &Logger{}

	if err := os.Remove(workingDirectory + "logs1.log"); err != nil {
		return err
	}

	if err := os.Rename(workingDirectory+"logs2.log", workingDirectory+"logs1.log"); err != nil {
		return err
	}

	if err := os.Rename(workingDirectory+"logs3.log", workingDirectory+"logs2.log"); err != nil {
		return err
	}

	file, err := setUpFile("logs3.log")
	if err != nil {
		return err
	}

	logger.file = file
	return nil
}

func (l *Logger) Close(ctx context.Context) {
	err := l.file.Close()
	if err != nil {
		log.Debug(ctx, err.Error())
	}
}

func (l *Logger) LogF(pattern string, args ...any) *LastMessage {
	msg := fmt.Sprintf(pattern, args...)

	_, _ = l.file.WriteString(msg + "\n")

	return (*LastMessage)(&msg)
}

func (lm *LastMessage) InConsole(ctx context.Context) {
	log.Debug(ctx, string(*lm))
}

func setUpFile(fileName string) (*os.File, error) {
	file, err := os.OpenFile(fmt.Sprintf("%s%s", workingDirectory, fileName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return file, nil
}

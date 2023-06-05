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
	workingDirectory = "./sarama/consumer/logs/"
)

// A little info can be found in segmentio/consumer/logger
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

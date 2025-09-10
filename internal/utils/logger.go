package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEUBG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type LogEvent struct {
	Level     LogLevel
	Message   string
	TimeStamp time.Time
}

type Logger struct {
	events chan LogEvent
	done   chan struct{}
	file   *os.File
	logger *log.Logger
}

func NewLogger() (*Logger, error) {
	logFile := viper.GetString(LogFile)
	local := viper.GetBool(Local)
	flag := os.O_CREATE | os.O_WRONLY
	if local {
		flag |= os.O_TRUNC
	} else {
		flag |= os.O_APPEND
	}

	file, err := os.OpenFile(logFile, flag, 0644)
	if err != nil {
		return nil, err
	}
	logger := log.New(file, "", log.LstdFlags|log.Lmicroseconds)

	return &Logger{
		events: make(chan LogEvent, 100),
		done:   make(chan struct{}),
		file:   file,
		logger: logger,
	}, nil
}

func (lg *Logger) Start() {
	go func() {
		defer close(lg.done)

		for ev := range lg.events {
			lg.logger.Printf("%s", formatEvent(ev))
		}
	}()
}

func (lg *Logger) log(level LogLevel, msg string) {
	select {
	case lg.events <- LogEvent{Level: level, Message: msg, TimeStamp: time.Now()}:
	default:
	}
}

func (lg *Logger) Stop() {
	close(lg.events)
	<-lg.done
	_ = lg.file.Close()
}

func (lg *Logger) Debug(component, msg string) {
	lg.log(DEBUG, fmt.Sprintf("[%s] %s", component, msg))
}

func (lg *Logger) Info(component, msg string) {
	lg.log(INFO, fmt.Sprintf("[%s] %s", component, msg))
}

func (lg *Logger) Warn(component, msg string) {
	lg.log(WARN, fmt.Sprintf("[%s] %s", component, msg))
}

func (lg *Logger) Error(component, msg string) {
	lg.log(ERROR, fmt.Sprintf("[%s] %s", component, msg))
}

func (lg *Logger) Fatal(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	lg.log(ERROR, formatted)
	panic(formatted)
}

func formatEvent(ev LogEvent) string {
	return fmt.Sprintf("[%s] %s - %s",
		ev.TimeStamp.Format(time.RFC3339),
		ev.Level,
		ev.Message,
	)
}

package logging

import (
	"errors"
	"fmt"
	"github.com/gookit/config/v2"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mkorman9/go-commons/logging/gelf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Setup(opts ...LoggingOpt) {
	loggingConfig := loggingConfig{}
	for _, opt := range opts {
		opt(&loggingConfig)
	}

	levelValue := config.String("logging.level")
	if levelValue == "" {
		levelValue = "info"
	}

	level, err := zerolog.ParseLevel(levelValue)
	if err == nil {
		zerolog.SetGlobalLevel(level)
	}

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.TimestampFieldName = "time"
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true
	zerolog.ErrorStackMarshaler = stackTraceMarshaller

	// configure default logger
	defaultWriter, _ := createWriter(
		&LoggerConfig{
			Output: "console",
			Format: "text",
			Text:   &TextConfig{Colors: true},
		},
	)
	log.Logger = log.Output(defaultWriter)

	// try to resolve loggers from configuration
	var loggers []*LoggerConfig
	err = config.BindStruct("logging.loggers", &loggers)

	var writers []io.Writer
	for _, loggerOpts := range loggers {
		writer, err := createWriter(loggerOpts)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to configure log writer")
		} else {
			writers = append(writers, writer)
		}
	}

	if len(writers) != 0 {
		log.Logger = log.Output(zerolog.MultiLevelWriter(writers...)) //.With().Str("hello", "world").Logger()
	}

	if len(loggingConfig.fields) != 0 {
		ctx := log.Logger.With()

		for name, value := range loggingConfig.fields {
			ctx = ctx.Str(name, value)
		}

		log.Logger = ctx.Logger()
	}
}

func createWriter(logger *LoggerConfig) (io.Writer, error) {
	logOutput, err := configureOutput(logger)
	if err != nil {
		return nil, err
	}

	writer, err := configureFormat(logger, logOutput)
	if err != nil {
		return nil, err
	}

	return writer, nil
}

func configureOutput(config *LoggerConfig) (io.Writer, error) {
	if config.Output == "gelf" {
		if config.Gelf == nil || len(config.Gelf.Address) == 0 {
			return nil, errors.New("logging output set to gelf but not properly configued")
		}

		gelfWriter, err := gelf.NewWriter(config.Gelf.Address)
		if err != nil {
			return nil, err
		}

		return gelfWriter, nil
	} else if config.Output == "file" {
		if config.File == nil {
			return nil, errors.New("logging output set to file but not properly configued")
		}

		fileWriter, err := os.OpenFile(config.File.Location, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}

		return fileWriter, nil
	} else if config.Output == "console" {
		return os.Stderr, nil
	} else {
		return nil, fmt.Errorf("unknown logging output: %v", config.Output)
	}
}

func configureFormat(config *LoggerConfig, output io.Writer) (io.Writer, error) {
	if config.Format == "text" {
		formattedOutput := zerolog.ConsoleWriter{
			Out:        output,
			NoColor:    !config.Text.Colors,
			TimeFormat: "2006-01-02 15:04:05",
		}

		return &formattedOutput, nil
	} else if config.Format == "json" {
		return output, nil
	} else {
		return nil, fmt.Errorf("unknown logging format: %v", config.Format)
	}
}

func stackTraceMarshaller(err error) interface{} {
	var stackTrace []string

	for i := 3; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)

		stackTrace = append(stackTrace, fmt.Sprintf("%v() [%v:%v]", fn.Name(), file, line))
	}

	return strings.Join(stackTrace, ", ")
}

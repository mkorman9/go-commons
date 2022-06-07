package logging

import (
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
	loggingConfig := loggingConfig{
		console: consoleConfig{
			enabled: true,
			colors:  true,
			format:  "text",
		},
		file: fileConfig{
			enabled:  false,
			location: "log.txt",
			format:   "text",
		},
		gelf: gelfConfig{
			enabled: true,
			address: "localhost:12201",
		},
	}
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

	if config.Exists("logging.console") {
		loggingConfig.console = consoleConfig{
			enabled: config.Bool("logging.console.enabled") || !config.Exists("logging.console.enabled"),
			colors:  config.Bool("logging.console.colors"),
			format:  config.String("logging.console.format"),
		}
	}
	if config.Exists("logging.file") {
		loggingConfig.file = fileConfig{
			enabled:  config.Bool("logging.file.enabled"),
			location: config.String("logging.file.location"),
			format:   config.String("logging.file.format"),
		}
	}
	if config.Exists("logging.gelf") {
		loggingConfig.gelf = gelfConfig{
			enabled: config.Bool("logging.gelf.enabled"),
			address: config.String("logging.gelf.address"),
		}
	}

	// default logger
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		NoColor:    false,
		TimeFormat: "2006-01-02 15:04:05",
	})

	var writers []io.Writer
	if loggingConfig.console.enabled {
		writer, err := configureFormat(os.Stderr, loggingConfig.console.format, loggingConfig.console.colors)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to configure console logger")
			return
		}

		writers = append(writers, writer)
	}
	if loggingConfig.file.enabled {
		fileWriter, err := os.OpenFile(loggingConfig.file.location, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to open file logger location")
			return
		}

		writer, err := configureFormat(fileWriter, loggingConfig.file.format, false)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to configure file logger")
			return
		}

		writers = append(writers, writer)
	}
	if loggingConfig.gelf.enabled {
		gelfWriter, err := gelf.NewWriter(loggingConfig.gelf.address)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create gelf logger connection")
			return
		}

		writer, err := configureFormat(gelfWriter, "json", false)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to configure gelf logger")
			return
		}

		writers = append(writers, writer)
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

func configureFormat(output io.Writer, format string, colors bool) (io.Writer, error) {
	if format == "text" {
		formattedOutput := zerolog.ConsoleWriter{
			Out:        output,
			NoColor:    !colors,
			TimeFormat: "2006-01-02 15:04:05",
		}

		return &formattedOutput, nil
	} else if format == "json" {
		return output, nil
	} else {
		return nil, fmt.Errorf("unknown logging format: %v", format)
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

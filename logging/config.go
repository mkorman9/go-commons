package logging

type loggingConfig struct {
	console consoleConfig
	file    fileConfig
	gelf    gelfConfig
	fields  map[string]string
}

type consoleConfig struct {
	enabled bool
	colors  bool
	format  string
}

type fileConfig struct {
	enabled  bool
	location string
	format   string
}

type gelfConfig struct {
	enabled bool
	address string
}

type LoggingOpt func(*loggingConfig)

func Fields(fields map[string]string) LoggingOpt {
	return func(config *loggingConfig) {
		config.fields = fields

		for key, value := range config.fields {
			config.fields[key] = value
		}
	}
}

func Field(name, value string) LoggingOpt {
	return func(config *loggingConfig) {
		if config.fields == nil {
			config.fields = make(map[string]string)
		}

		config.fields[name] = value
	}
}

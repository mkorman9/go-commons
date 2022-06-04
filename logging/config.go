package logging

type loggingConfig struct {
	fields map[string]string
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

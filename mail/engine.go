package mail

import (
	"errors"
	"github.com/gookit/config/v2"
)

type Engine struct {
	backend Backend
}

func NewEngine() (*Engine, error) {
	name := config.String("mail.engine.name")
	sourceName := config.String("mail.engine.source.name")
	sourceEmail := config.String("mail.engine.source.email")

	if name == "" {
		name = "fake"
	}

	if sourceName == "" || sourceEmail == "" {
		return nil, errors.New("source name and source email cannot be empty")
	}

	var backend Backend
	switch name {
	case "fake":
		backend = &fakeBackend{sourceName: sourceName, sourceEmail: sourceEmail}
	case "sendgrid":
		b, err := newSendgridBackend(sourceName, sourceEmail)
		if err != nil {
			return nil, err
		}

		backend = b
	default:
		return nil, errors.New("unknown backend")
	}

	return &Engine{
		backend: backend,
	}, nil
}

func (engine *Engine) SendMail(targetName, targetEmail, template string, props Props) error {
	return engine.backend.SendMail(targetName, targetEmail, template, props)
}

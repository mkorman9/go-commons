package mail

import (
	"errors"
	"github.com/gookit/config/v2"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type sendgridBackend struct {
	client *sendgrid.Client

	sourceName  string
	sourceEmail string
}

func newSendgridBackend(sourceName, sourceEmail string) (*sendgridBackend, error) {
	token := config.String("mail.sendgrid.token")

	if token == "" {
		return nil, errors.New("sendgrid token cannot be empty")
	}

	return &sendgridBackend{
		client:      sendgrid.NewSendClient(token),
		sourceName:  sourceName,
		sourceEmail: sourceEmail,
	}, nil
}

func (backend *sendgridBackend) SendMail(targetName, targetEmail, template string, props Props) error {
	from := mail.NewEmail(backend.sourceName, backend.sourceEmail)
	to := mail.NewEmail(targetName, targetEmail)

	m := mail.NewV3Mail()
	m.SetFrom(from)
	m.SetTemplateID(template)

	personalization := mail.NewPersonalization()
	personalization.AddTos(to)

	for key, value := range props {
		personalization.SetDynamicTemplateData(key, value)
	}

	_, err := backend.client.Send(m)
	return err
}

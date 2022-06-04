package mail

import "github.com/rs/zerolog/log"

type fakeBackend struct {
	sourceName  string
	sourceEmail string
}

func (backend *fakeBackend) SendMail(targetName, targetEmail, template string, props Props) error {
	log.Info().Msgf(
		"Sending fake E-Mail from='%s (%s)', to='%s (%s), template=%s, props=%v'",
		backend.sourceName,
		backend.sourceEmail,
		targetName,
		targetEmail,
		template,
		props,
	)

	return nil
}

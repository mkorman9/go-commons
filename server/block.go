package server

import (
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

var shutdownSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

func BlockThread(errorChannel <-chan error) {
	shutdownSignalsChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownSignalsChannel, shutdownSignals...)

	for {
		select {
		case err := <-errorChannel:
			log.Error().Err(err).Msg("Unblocking thread due to an error")
			return
		case s := <-shutdownSignalsChannel:
			log.Info().Msgf("Exiting due to a signal: %v", s)
			return
		}
	}
}

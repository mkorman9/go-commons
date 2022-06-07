package firestorelib

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/gookit/config/v2"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"os"
	"time"
)

var defaultConnectTimeout = 5 * time.Second

func NewClient() (*firestore.Client, func(), error) {
	projectID := config.String("gcp.projectId")
	emulatorEnabled := config.Bool("gcp.firestore.emulator.enabled")
	emulatorAddress := config.String("gcp.firestore.emulator.address")
	connectTimeoutValue := config.Int64("gcp.firestore.timeouts.connect")

	options := []option.ClientOption{
		option.WithGRPCDialOption(grpc.WithReturnConnectionError()),
	}

	if projectID == "" {
		log.Warn().Msg("Empty gcp.projectId, using default")
		projectID = "default-project-id"
	}

	if emulatorEnabled {
		if emulatorAddress == "" {
			emulatorAddress = "127.0.0.1:8538"
		}

		if err := os.Setenv("FIRESTORE_EMULATOR_HOST", emulatorAddress); err != nil {
			return nil, func() {}, err
		}
	} else {
		options = append(options, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	}

	connectTimeout := defaultConnectTimeout
	if connectTimeoutValue != 0 {
		connectTimeout = time.Duration(connectTimeoutValue) * time.Millisecond
	}

	log.Debug().Msg("Establishing Firestore connection")

	connectContext, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	fs, err := firestore.NewClient(connectContext, projectID, options...)
	if err != nil {
		return nil, func() {}, err
	}

	log.Info().Msg("Successfully connected to Firestore")

	return fs, func() { closeClient(fs) }, nil
}

func closeClient(fs *firestore.Client) {
	log.Debug().Msg("Closing Firestore connection")

	err := fs.Close()
	if err == nil {
		log.Info().Msg("Firestore connection closed successfully")
	} else {
		log.Error().Err(err).Msg("Error while closing Firestore connection")
	}
}

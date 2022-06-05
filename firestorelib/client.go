package firestorelib

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/gookit/config/v2"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"os"
	"time"
)

var defaultConnectTimeout = 5 * time.Second

type Client struct {
	fs *firestore.Client
}

func NewClient() (*Client, error) {
	projectID := config.String("gcp.projectId")
	emulatorEnabled := config.Bool("gcp.firestore.emulator.enabled")
	emulatorAddress := config.String("gcp.firestore.emulator.address")
	connectTimeoutValue := config.Int64("gcp.firestore.timeouts.connect")

	var options []option.ClientOption

	if projectID == "" {
		projectID = os.Getenv("GOOGLE_PROJECT_ID")

		if projectID == "" {
			log.Warn().Msg("Empty gcp.projectId, using default")
			projectID = "default-project-id"
		}
	}

	if emulatorEnabled {
		if emulatorAddress == "" {
			emulatorAddress = "127.0.0.1:8538"
		}

		if err := os.Setenv("FIRESTORE_EMULATOR_HOST", emulatorAddress); err != nil {
			return nil, err
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

	client, err := firestore.NewClient(connectContext, projectID, options...)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Successfully connected to Firestore")

	return &Client{
		fs: client,
	}, nil
}

func (client *Client) Close() {
	log.Debug().Msg("Closing Firestore connection")

	err := client.fs.Close()
	if err == nil {
		log.Info().Msg("Firestore connection closed successfully")
	} else {
		log.Error().Err(err).Msg("Error while closing Firestore connection")
	}
}

func (client *Client) Get() *firestore.Client {
	return client.fs
}

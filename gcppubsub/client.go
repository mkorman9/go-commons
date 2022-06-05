package gcppubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/googleapis/gax-go/v2/apierror"
	"github.com/gookit/config/v2"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/option"
	"os"
	"time"
)

var defaultConnectTimeout = 5 * time.Second
var defaultTopicCreateTimeout = 3 * time.Second
var defaultSubscriptionCreateTimeout = 3 * time.Second
var defaultSubscriptionDeleteTimeout = 8 * time.Second
var defaultPublishTimeout = 3 * time.Second

type MessageChannel <-chan *pubsub.Message

type Client struct {
	client *pubsub.Client

	topicCreateTimeout        time.Duration
	subscriptionCreateTimeout time.Duration
	subscriptionDeleteTimeout time.Duration
	publishTimeout            time.Duration

	receivesToCancel       []func()
	ephemeralSubscriptions []*pubsub.Subscription
	topics                 map[string]*pubsub.Topic
}

func NewClient() (*Client, error) {
	projectID := config.String("gcp.projectId")
	emulatorEnabled := config.Bool("gcp.pubsub.emulator.enabled")
	emulatorAddress := config.String("gcp.pubsub.emulator.address")
	connectTimeoutValue := config.Int64("gcp.pubsub.timeouts.connect")
	topicCreateTimeoutValue := config.Int64("gcp.pubsub.timeouts.topicCreate")
	subscriptionCreateTimeoutValue := config.Int64("gcp.pubsub.timeouts.subscriptionCreate")
	subscriptionDeleteTimeoutValue := config.Int64("gcp.pubsub.timeouts.subscriptionDelete")
	publishTimeoutValue := config.Int64("gcp.pubsub.timeouts.publish")

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

		if err := os.Setenv("PUBSUB_EMULATOR_HOST", emulatorAddress); err != nil {
			return nil, err
		}
	} else {
		options = append(options, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	}

	var connectTimeout = defaultConnectTimeout
	if connectTimeoutValue != 0 {
		connectTimeout = time.Duration(connectTimeoutValue) * time.Millisecond
	}

	var topicCreateTimeout = defaultTopicCreateTimeout
	if topicCreateTimeoutValue != 0 {
		topicCreateTimeout = time.Duration(topicCreateTimeoutValue) * time.Millisecond
	}

	var subscriptionCreateTimeout = defaultSubscriptionCreateTimeout
	if subscriptionCreateTimeoutValue != 0 {
		subscriptionCreateTimeout = time.Duration(subscriptionCreateTimeoutValue) * time.Millisecond
	}

	var subscriptionDeleteTimeout = defaultSubscriptionDeleteTimeout
	if subscriptionDeleteTimeoutValue != 0 {
		subscriptionDeleteTimeout = time.Duration(subscriptionDeleteTimeoutValue) * time.Millisecond
	}

	var publishTimeout = defaultPublishTimeout
	if publishTimeoutValue != 0 {
		publishTimeout = time.Duration(publishTimeoutValue) * time.Millisecond
	}

	log.Debug().Msg("Establishing PubSub connection")

	connectContext, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	client, err := pubsub.NewClient(connectContext, projectID, options...)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Successfully connected to PubSub")

	return &Client{
		client:                    client,
		topicCreateTimeout:        topicCreateTimeout,
		subscriptionCreateTimeout: subscriptionCreateTimeout,
		subscriptionDeleteTimeout: subscriptionDeleteTimeout,
		publishTimeout:            publishTimeout,
		topics:                    make(map[string]*pubsub.Topic),
	}, nil
}

func (pubSubClient *Client) Close() {
	log.Debug().Msg("Closing PubSub connection")

	for _, cancel := range pubSubClient.receivesToCancel {
		cancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), pubSubClient.subscriptionDeleteTimeout)
	defer cancel()

	for _, subscription := range pubSubClient.ephemeralSubscriptions {
		err := subscription.Delete(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Error while deleting PubSub subscription")
		}
	}

	for _, topic := range pubSubClient.topics {
		topic.Stop()
	}

	err := pubSubClient.client.Close()
	if err == nil {
		log.Info().Msg("PubSub connection closed successfully")
	} else {
		log.Error().Err(err).Msg("Error while closing PubSub connection")
	}
}

func (pubSubClient *Client) CreatePublisher(topicName string) (*Publisher, error) {
	topic, err := pubSubClient.getTopic(topicName)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		topic:          topic,
		publishTimeout: pubSubClient.publishTimeout,
	}, nil
}

func (pubSubClient *Client) SubscribeToTopic(topicName string) (MessageChannel, error) {
	subscriptionName := fmt.Sprintf("%s.%s", topicName, uuid.NewV4().String())

	subscription, err := pubSubClient.getSubscription(subscriptionName, topicName)
	if err != nil {
		return nil, err
	}

	pubSubClient.ephemeralSubscriptions = append(pubSubClient.ephemeralSubscriptions, subscription)

	return pubSubClient.startReceive(subscription), nil
}

func (pubSubClient *Client) AttachToSubscription(subscriptionName, topicName string) (MessageChannel, error) {
	subscription, err := pubSubClient.getSubscription(subscriptionName, topicName)
	if err != nil {
		return nil, err
	}

	return pubSubClient.startReceive(subscription), nil
}

func (pubSubClient *Client) getTopic(name string) (*pubsub.Topic, error) {
	if topic, ok := pubSubClient.topics[name]; ok {
		return topic, nil
	}

	createContext, cancel := context.WithTimeout(context.Background(), pubSubClient.topicCreateTimeout)
	defer cancel()

	topic, err := pubSubClient.client.CreateTopic(createContext, name)
	if err != nil {
		handled := false

		if apiErr, ok := err.(*apierror.APIError); ok {
			if apiErr.GRPCStatus().Message() == "Topic already exists" {
				topic = pubSubClient.client.Topic(name)
				handled = true
			}
		}

		if !handled {
			return nil, err
		}
	}

	pubSubClient.topics[name] = topic

	return topic, nil
}

func (pubSubClient *Client) getSubscription(
	subscriptionName,
	topicName string,
) (*pubsub.Subscription, error) {
	topic, err := pubSubClient.getTopic(topicName)
	if err != nil {
		return nil, err
	}

	createContext, cancel := context.WithTimeout(context.Background(), pubSubClient.subscriptionCreateTimeout)
	defer cancel()

	subscription, err := pubSubClient.client.CreateSubscription(
		createContext,
		subscriptionName,
		pubsub.SubscriptionConfig{Topic: topic},
	)
	if err != nil {
		handled := false

		if apiErr, ok := err.(*apierror.APIError); ok {
			if apiErr.GRPCStatus().Message() == "Subscription already exists" {
				subscription = pubSubClient.client.Subscription(subscriptionName)
				handled = true
			}
		}

		if !handled {
			return nil, err
		}
	}

	return subscription, nil
}

func (pubSubClient *Client) startReceive(subscription *pubsub.Subscription) MessageChannel {
	ctx, cancel := context.WithCancel(context.Background())
	pubSubClient.receivesToCancel = append(pubSubClient.receivesToCancel, cancel)

	messageChannel := make(chan *pubsub.Message)

	go func() {
		_ = subscription.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
			messageChannel <- msg
		})
	}()

	return messageChannel
}

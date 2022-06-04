package gcppubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"time"
)

type Publisher struct {
	topic          *pubsub.Topic
	publishTimeout time.Duration
}

func (publisher *Publisher) PublishAsync(msg interface{}) error {
	marshaledMessage, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), publisher.publishTimeout)
	defer cancel()

	_ = publisher.topic.Publish(ctx, &pubsub.Message{Data: marshaledMessage})

	return nil
}

func (publisher *Publisher) Publish(msg interface{}) error {
	marshaledMessage, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), publisher.publishTimeout)
	defer cancel()

	result := publisher.topic.Publish(ctx, &pubsub.Message{Data: marshaledMessage})

	_, err = result.Get(ctx)
	if err != nil {
		return err
	}

	return nil
}

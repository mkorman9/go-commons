package redislib

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/gookit/config/v2"
	"github.com/rs/zerolog/log"
	"time"
)

type RedisClient struct {
	client *redis.Client
}

func NewConnection() (*RedisClient, error) {
	address := config.String("redis.address")
	username := config.String("redis.username")
	password := config.String("redis.password")
	db := config.Int("redis.db")
	enableTLS := config.Bool("redis.tls")
	connectionTimeoutValue := config.Int64("redis.timeouts.connection")

	if address == "" {
		return nil, errors.New("redis.address cannot be empty")
	}

	connectionTimeout := 5 * time.Second
	if connectionTimeoutValue > 0 {
		connectionTimeout = time.Duration(connectionTimeoutValue) * time.Millisecond
	}

	options := redis.Options{
		Addr:     address,
		Username: username,
		Password: password,
		DB:       db,
	}

	if enableTLS {
		options.TLSConfig = &tls.Config{}
	}

	client := redis.NewClient(&options)

	log.Debug().Msg("Establishing Redis connection")

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	} else {
		log.Info().Msg("Successfully connected to Redis")
	}

	return &RedisClient{client}, nil
}

func (c *RedisClient) Get() *redis.Client {
	return c.client
}

func (c *RedisClient) Close() {
	log.Debug().Msg("Closing Redis connection")

	if err := c.client.Close(); err != nil {
		log.Error().Err(err).Msg("Error when closing Redis connection")
	} else {
		log.Info().Msg("Redis connection closed successfully")
	}
}

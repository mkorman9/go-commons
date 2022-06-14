package tcpserver

import (
	"context"
	"errors"
	"github.com/gookit/config/v2"
	"github.com/rs/zerolog/log"
	"net"
)

type Handler = func(ctx context.Context, connection net.Conn)

type Server struct {
	address string

	listener net.Listener
	handler  Handler
}

func NewServer() *Server {
	address := config.String("server.tcp.address")

	if address == "" {
		address = "0.0.0.0:5000"
	}

	return &Server{
		address: address,
	}
}

func (server *Server) Handler(h Handler) {
	server.handler = h
}

func (server *Server) Start(errorChannel chan<- error) {
	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		errorChannel <- err
		return
	}

	server.listener = listener

	log.Info().Msgf("Started TCP server on %v", server.address)

	go func() {
		ctx, cancel := context.WithCancel(context.Background())

		for {
			connection, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					cancel()
					break // server is being shut down
				}

				log.Error().Err(err).Msg("Error while accepting TCP connection")
				continue
			}

			if server.handler != nil {
				go server.handler(ctx, connection)
			}
		}
	}()
}

func (server *Server) Stop() {
	if server.listener != nil {
		log.Debug().Msg("Shutting down TCP server")

		err := server.listener.Close()
		server.listener = nil

		if err != nil {
			log.Error().Err(err).Msg("Error shutting down TCP server")
		} else {
			log.Info().Msg("TCP server shutdown successful")
		}
	}
}

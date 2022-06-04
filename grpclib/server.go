package grpclib

import (
	"github.com/gookit/config/v2"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	server  *grpc.Server
	address string
}

func NewServer(options ...grpc.ServerOption) *Server {
	address := config.String("grpc.address")

	if address == "" {
		address = "0.0.0.0:9090"
	}

	return &Server{
		server:  grpc.NewServer(options...),
		address: address,
	}
}

func (s *Server) Get() *grpc.Server {
	return s.server
}

func (s *Server) Start(errorChannel chan<- error) {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		errorChannel <- err
		return
	}

	log.Info().Msgf("Started gRPC server on %s", s.address)

	go func() {
		if err := s.server.Serve(listener); err != nil {
			errorChannel <- err
		}
	}()
}

func (s *Server) Stop() {
	log.Info().Msg("Stopping gRPC server")
	s.server.GracefulStop()
}

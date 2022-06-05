package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gookit/config/v2"
	"github.com/mkorman9/go-commons/web"
	"github.com/rs/zerolog/log"
	"net/http"
)

type Server struct {
	Engine     *gin.Engine
	HttpServer *http.Server

	address string
}

func NewServer() *Server {
	address := config.String("server.address")
	trustedProxies := config.Strings("server.trustedProxies")

	if address == "" {
		address = "0.0.0.0:8080"
	}

	if trustedProxies == nil {
		trustedProxies = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.1/8"}
	}

	engine := createEngine(trustedProxies)

	return &Server{
		Engine: engine,
		HttpServer: &http.Server{
			Addr:    address,
			Handler: engine,
		},
		address: address,
	}
}

func (server *Server) Start(errorChannel chan<- error) {
	go server.runServer(errorChannel)
}

func (server *Server) Stop() {
	log.Debug().Msg("Shutting down HTTP server")

	err := server.HttpServer.Shutdown(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Error shutting down http server")
	} else {
		log.Info().Msg("Server shutdown successful")
	}
}

func (server *Server) runServer(errorChannel chan<- error) {
	log.Info().Msgf("Started HTTP server on %v", server.address)

	err := server.HttpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		errorChannel <- err
	}
}

func createEngine(trustedProxies []string) *gin.Engine {
	engine := gin.New()

	engine.Use(recoveryMiddleware())

	engine.ForwardedByClientIP = true
	_ = engine.SetTrustedProxies(trustedProxies)

	engine.HandleMethodNotAllowed = true
	engine.NoMethod(func(c *gin.Context) {
		web.ErrorResponse(c, http.StatusMethodNotAllowed, "Method not allowed")
	})

	engine.NoRoute(func(c *gin.Context) {
		web.ErrorResponse(c, http.StatusNotFound, "Not found")
	})

	return engine
}

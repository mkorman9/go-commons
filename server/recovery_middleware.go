package server

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"os"
	"syscall"
)

func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				if isPipeWriteError(r) { // client has closed the connection while server was sending response
					return
				}

				log.Error().Stack().Err(fmt.Errorf("%v", r)).Msg("Panic inside a handler function")
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}

func isPipeWriteError(r interface{}) bool {
	if opErr, ok := r.(*net.OpError); ok && opErr.Op == "write" {
		if syscallError, ok := opErr.Err.(*os.SyscallError); ok {
			return errors.Is(syscallError.Err, syscall.EPIPE)
		}
	}

	return false
}

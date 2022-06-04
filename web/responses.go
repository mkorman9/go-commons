package web

import (
	"github.com/rs/zerolog/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Cause struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

type GenericResponse struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	Causes  []Cause `json:"causes,omitempty"`
}

func SuccessResponse(c *gin.Context, message string) {
	response(c, http.StatusOK, "success", message)
}

func ErrorResponse(c *gin.Context, code int, message string, causes ...Cause) {
	response(c, code, "error", message, causes...)
}

func InternalError(c *gin.Context, err error, logMessage string, v ...interface{}) {
	log.Error().Err(err).Msgf(logMessage, v...)
	ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
}

func FieldError(field, code string) Cause {
	return Cause{Field: field, Code: code}
}

func FieldErrorMessage(field, code, message string) Cause {
	return Cause{Field: field, Code: code, Message: message}
}

func response(c *gin.Context, code int, status, message string, causes ...Cause) {
	ca := causes
	if ca == nil {
		ca = make([]Cause, 0)
	}

	c.JSON(code, &GenericResponse{Status: status, Message: message, Causes: ca})
}

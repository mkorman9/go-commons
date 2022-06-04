package httpauth

import (
	"github.com/gin-gonic/gin"
	"github.com/mkorman9/go-commons/web"
	"net/http"
	"strings"
)

type VerifyTokenFunc = func(c *gin.Context, token string) (*VerificationResult, error)

func NewBearerTokenMiddleware(verifyToken VerifyTokenFunc) Middleware {
	return newMiddleware(
		func(rolesCheckingFunc RolesCheckingFunc) gin.HandlerFunc {
			return func(c *gin.Context) {
				token := extractToken(c)

				verificationResult, err := verifyToken(c, token)
				if err != nil {
					web.InternalError(c, err, "Error while trying to verify token")
					return
				}

				if !verificationResult.Verified {
					c.AbortWithStatusJSON(
						http.StatusUnauthorized,
						&web.GenericResponse{
							Status:  "error",
							Message: "Invalid token",
							Causes: []web.Cause{
								web.FieldErrorMessage(
									"token",
									"unverified",
									"Token cannot be verified",
								),
							},
						},
					)
					return
				}

				rolesCheckingResult := rolesCheckingFunc(verificationResult.Roles)

				if !rolesCheckingResult {
					c.AbortWithStatusJSON(
						http.StatusForbidden,
						&web.GenericResponse{
							Status:  "error",
							Message: "Access Denied",
							Causes: []web.Cause{
								web.FieldErrorMessage(
									"token",
									"unauthorized",
									"Token does not grant the role required to access",
								),
							},
						},
					)
					return
				}

				c.Next()
			}
		},
	)
}

func extractToken(c *gin.Context) string {
	authorizationHeader := c.GetHeader("Authorization")
	if len(authorizationHeader) == 0 {
		return ""
	}

	fields := strings.Fields(authorizationHeader)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return ""
	}

	token := fields[1]
	if len(token) == 0 {
		return ""
	}

	return token
}

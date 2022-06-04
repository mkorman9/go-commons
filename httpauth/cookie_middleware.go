package httpauth

import (
	"github.com/gin-gonic/gin"
	"github.com/mkorman9/go-commons/web"
	"net/http"
)

type VerifyCookieFunc = func(c *gin.Context, cookie string) (*VerificationResult, error)

func NewSessionCookieMiddleware(cookieName string, verifyCookie VerifyCookieFunc) Middleware {
	return newMiddleware(
		func(rolesCheckingFunc RolesCheckingFunc) gin.HandlerFunc {
			return func(c *gin.Context) {
				cookie, err := c.Cookie(cookieName)
				if err != nil {
					cookie = ""
				}

				verificationResult, err := verifyCookie(c, cookie)
				if err != nil {
					web.InternalError(c, err, "Error while trying to verify cookie")
					return
				}

				if !verificationResult.Verified {
					c.AbortWithStatusJSON(
						http.StatusUnauthorized,
						&web.GenericResponse{
							Status:  "error",
							Message: "Invalid session cookie",
							Causes: []web.Cause{
								web.FieldErrorMessage(
									"cookie",
									"unverified",
									"Session cookie cannot be verified",
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
									"Session cookie does not grant the role required to access",
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

package httpauth

import "github.com/gin-gonic/gin"

type VerificationResult struct {
	Verified bool
	Roles    []string
}

type RolesCheckingFunc = func(roles []string) bool
type MiddlewareHandler = func(rolesCheckingFunc RolesCheckingFunc) gin.HandlerFunc

type Middleware struct {
	handler MiddlewareHandler
}

func newMiddleware(handler MiddlewareHandler) Middleware {
	return Middleware{handler}
}

func (middleware *Middleware) Anyone() gin.HandlerFunc {
	return middleware.handler(func(_ []string) bool {
		return true
	})
}

func (middleware *Middleware) AnyAuthenticated() gin.HandlerFunc {
	return middleware.handler(func(_ []string) bool {
		return true
	})
}

func (middleware *Middleware) AnyOfRoles(allowedRoles ...string) gin.HandlerFunc {
	allowedRolesSet := make(map[string]struct{})
	for _, role := range allowedRoles {
		allowedRolesSet[role] = struct{}{}
	}

	return middleware.handler(func(providedRoles []string) bool {
		hasRole := false
		for _, role := range providedRoles {
			if _, ok := allowedRolesSet[role]; ok {
				hasRole = true
				break
			}
		}

		return hasRole
	})
}

func (middleware *Middleware) AllOfRoles(requiredRoles ...string) gin.HandlerFunc {
	return middleware.handler(func(providedRoles []string) bool {
		for _, role := range requiredRoles {
			hasRole := false
			for _, providedRole := range providedRoles {
				if role == providedRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				return false
			}
		}

		return true
	})
}

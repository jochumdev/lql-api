package http

import (
	"net/http"
	"strconv"

	auth "github.com/abbot/go-http-auth"
	"github.com/gin-gonic/gin"
)

// BasicAuthMiddleware gin middleware
func BasicAuthMiddleware(a *auth.BasicAuth) gin.HandlerFunc {
	realmHeader := "Basic realm=" + strconv.Quote(a.Realm)

	return func(c *gin.Context) {
		user := a.CheckAuth(c.Request)

		if user == "" {
			c.Header("WWW-Authenticate", realmHeader)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", user)
	}
}

package lql

import (
	"errors"
	"fmt"
	"net/http"

	auth "github.com/abbot/go-http-auth"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	// Logger is the Global logrus logger
	Logger *log.Logger
)

// CtxKeyLQLClient is they key for the Client in the gin context.
const CtxKeyLQLClient = "lqlClient"

// GinGetLqlClient gets the LQL Client from a gin context
func GinGetLqlClient(c *gin.Context) (*Client, error) {
	clientIface, ok := c.Get(CtxKeyLQLClient)
	if !ok {
		return nil, errors.New("Failed to get the LQL client from context")
	}

	return clientIface.(*Client), nil
}

func clientInjectorMiddleware(client *Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(CtxKeyLQLClient, client)
	}
}

func basicAuthMiddleware(htpasswdPath, realm string) gin.HandlerFunc {
	htpasswd := auth.HtpasswdFileProvider(htpasswdPath)
	authenticator := auth.NewBasicAuthenticator(realm, htpasswd)
	realmHeader := fmt.Sprintf("Basic realm=\"%s\"", realm)

	return func(c *gin.Context) {
		user := authenticator.CheckAuth(c.Request)

		if user == "" {
			c.Header("WWW-Authenticate", realmHeader)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", user)
	}
}

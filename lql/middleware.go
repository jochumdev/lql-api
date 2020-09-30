package lql

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	auth "github.com/abbot/go-http-auth"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
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

func basicAuthWithWatcherMiddleware(htpasswdPath, realm string) gin.HandlerFunc {
	authenticatorLock := sync.RWMutex{}
	htpasswd := auth.HtpasswdFileProvider(htpasswdPath)
	authenticator := auth.NewBasicAuthenticator(realm, htpasswd)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				Logger.WithField("event", event).Debug("event")
				if event.Op&fsnotify.Write == fsnotify.Write {
					authenticatorLock.Lock()
					Logger.WithField("path", event.Name).Debug("Modified file")
					htpasswd = auth.HtpasswdFileProvider(htpasswdPath)
					authenticator = auth.NewBasicAuthenticator(realm, htpasswd)
					authenticatorLock.Unlock()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				Logger.WithField("error", err).Error()
			}
		}
	}()

	err = watcher.Add(htpasswdPath)
	if err != nil {
		log.Fatal(err)
	}

	realmHeader := fmt.Sprintf("Basic realm=\"%s\"", realm)
	return func(c *gin.Context) {
		authenticatorLock.RLock()
		user := authenticator.CheckAuth(c.Request)
		authenticatorLock.RUnlock()

		if user == "" {
			c.Header("WWW-Authenticate", realmHeader)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", user)
	}
}

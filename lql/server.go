package lql

import (
	"fmt"
	"net/http"

	auth "github.com/abbot/go-http-auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"

	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
)

type Server struct {
	fizz         *fizz.Fizz
	htpasswdPath string
}

func NewServer(client *Client, logger *log.Logger, htpasswdPath string) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(cors.Default())
	engine.Use(ginlogrus.Logger(logger), gin.Recovery(), clientInjectorMiddleware(client))

	Logger = logger

	fizz := fizz.NewFromEngine(engine)

	// Override type names.
	// fizz.Generator().OverrideTypeName(reflect.TypeOf(Fruit{}), "SweetFruit")

	// Initialize the informations of
	// the API that will be served with
	// the specification.
	infos := &openapi.Info{
		Title: "LQL API",
		Description: `This is the LQL API for your check_mk Server.

All v1/ endpoints require http basic auth`,
		Version: "unset",
	}
	// Create a new route that serve the OpenAPI spec.
	fizz.GET("/openapi.json", nil, fizz.OpenAPI(infos, "json"))

	// Setup routes.
	v1Group := fizz.Group("/v1", "v1", "LQL API v1")
	if htpasswdPath != "" {
		htpasswd := auth.HtpasswdFileProvider(htpasswdPath)
		authenticator := auth.NewBasicAuthenticator("LQL API", htpasswd)
		v1Group.Use(basicAuthMiddleware(authenticator))
	} else {
		// Inject empty user if not .htpasswd have been given
		v1Group.Use(func(c *gin.Context) {
			c.Set("user", "")
		})
	}
	v1Routes(v1Group)

	// routes(fizz.Group("/market", "market", "Your daily dose of freshness"))

	if len(fizz.Errors()) != 0 {
		return nil, fmt.Errorf("fizz errors: %v", fizz.Errors())
	}
	return &Server{fizz: fizz, htpasswdPath: htpasswdPath}, nil
}

func (s *Server) ListenAndServe(address string) {
	srv := &http.Server{
		Addr:    address,
		Handler: s.fizz,
	}

	srv.ListenAndServe()
}

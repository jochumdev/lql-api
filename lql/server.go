package lql

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
)

type Server struct {
	client *Client
	fizz   *fizz.Fizz
}

func NewServer(client *Client) (*Server, error) {
	engine := gin.New()
	engine.Use(cors.Default())

	fizz := fizz.NewFromEngine(engine)

	// Override type names.
	// fizz.Generator().OverrideTypeName(reflect.TypeOf(Fruit{}), "SweetFruit")

	// Initialize the informations of
	// the API that will be served with
	// the specification.
	infos := &openapi.Info{
		Title:       "LQL API",
		Description: `This is the LQL API for your check_mk Server.`,
		Version:     "unset",
	}
	// Create a new route that serve the OpenAPI spec.
	fizz.GET("/openapi.json", nil, fizz.OpenAPI(infos, "json"))

	// Setup routes.
	// routes(fizz.Group("/market", "market", "Your daily dose of freshness"))

	if len(fizz.Errors()) != 0 {
		return nil, fmt.Errorf("fizz errors: %v", fizz.Errors())
	}
	return &Server{client: client, fizz: fizz}, nil
}

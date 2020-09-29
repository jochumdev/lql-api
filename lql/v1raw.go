package lql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// V1RawRequestParams is a request for the RAW API
type V1RawRequestParams struct {
	Method string `json:"method" validate:"required" description:"Either GET or COMMAND"`

	// COMMAND
	Command string `json:"command" description:"The command, required if method is COMMAND"`

	// GET
	Table   string     `json:"table" description:"The table to query, required if method is GET"`
	Columns []string   `json:"columns" description:"Columns to query, you should always provide this"`
	Query   [][]string `json:"query" description:"raw query Data"`

	// both
	Limit int `json:"limit" description:"Limit result count"`
}

// GetRaw get a raw request
func v1RawPost(c *gin.Context, params *V1RawRequestParams) ([]gin.H, error) {
	client, err := GinGetLqlClient(c)
	if err != nil {
		return nil, err
	}

	// Param validation and request building
	request := []string{}
	switch strings.ToLower(params.Method) {
	case "get":
		if params.Table == "" {
			return nil, errors.New("Param table is required with method 'GET'")
		}
		request = append(request, fmt.Sprintf("GET %s", params.Table))

		if len(params.Columns) > 0 {
			request = append(request, fmt.Sprintf("Columns: %s\n", strings.Join(params.Columns, " ")))
		}

		if len(params.Query) > 0 {
			for _, q := range params.Query {
				if len(q) != 2 {
					return nil, errors.New("Each query must contain a key and a value")
				}
				request = append(request, fmt.Sprintf("%s: %s", q[0], q[1]))
			}
		}
		break
	case "command":
		if params.Command == "" {
			return nil, errors.New("Param command is required with method 'COMMAND'")
		}
		request = append(request, fmt.Sprintf("COMMAND %s", params.Command))
		break
	default:
		return nil, fmt.Errorf("Unknown Method requested: '%s'", params.Method)
	}

	return client.Request(c, strings.Join(request, "\n"), "", params.Limit)
}

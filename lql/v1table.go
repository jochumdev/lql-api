package lql

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

var v1TableColumns = map[string][]string{}

func init() {
	v1TableColumns = make(map[string][]string, 1) // Increment this when you add tables
	v1TableColumns["hosts"] = []string{
		"name",
		"display_name",
		"address",
		"alias",
		"tags",
		"labels",
		"groups",
		"latency",
		"parents",
	}
}

type v1TableGetParams struct {
	Table   string   `path:"name"`
	Columns *string  `query:"columns" description:"Columns to return" validate:"omitempty"`
	Limit   *float64 `query:"limit" description:"Limit number of results" validate:"omitempty,min=0"`
}

func v1TableGet(c *gin.Context, params *v1TableGetParams) ([]gin.H, error) {
	client, err := GinGetLqlClient(c)
	if err != nil {
		return nil, err
	}
	user := c.GetString("user")

	columns := ""
	if params.Columns != nil {
		columns = strings.Join(strings.Split(*params.Columns, ","), " ")
	} else if defaultCols, ok := v1TableColumns[params.Table]; ok {
		columns = strings.Join(defaultCols, " ")
	} else {
		columns = "name"
	}

	limit := 0
	if params.Limit != nil {
		limit = int(*params.Limit)
	}

	lines := []string{fmt.Sprintf("GET %s", params.Table), fmt.Sprintf("Columns: %s", columns)}
	resp, err := client.Request(c, strings.Join(lines, "\n"), user, limit)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type v1TableGetColumnsParams struct {
	Table string `path:"name"`
}

func v1TableGetColumns(c *gin.Context, params *v1TableGetColumnsParams) ([]string, error) {
	client, err := GinGetLqlClient(c)
	if err != nil {
		return nil, err
	}
	user := c.GetString("user")

	msg := fmt.Sprintf("GET columns\nColumns: name\nFilter: table = %s", params.Table)
	resp, err := client.Request(c, msg, user, 0)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(resp))
	for i, item := range resp {
		result[i] = item["name"].(string)
	}

	return result, nil
}

package lql

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

var v1TableColumns = map[string][]string{}
var v1TableFilters = map[string][]string{}

func init() {
	v1TableColumns = make(map[string][]string, 2) // Increment this when you add tables
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
	v1TableColumns["services"] = []string{
		"state",
		"host_name",
		"display_name",
		"description",
		"plugin_output",
	}

	v1TableFilters = make(map[string][]string, 3)
	v1TableFilters["service_problems"] = []string{
		"Filter: state > 0",
		"Filter: scheduled_downtime_depth = 0",
		"Filter: host_scheduled_downtime_depth = 0",
		"Filter: host_state = 0",
	}
	v1TableFilters["service_unhandled"] = []string{
		"Filter: state > 0",
		"Filter: scheduled_downtime_depth = 0",
		"Filter: host_scheduled_downtime_depth = 0",
		"Filter: acknowledged = 0",
		"Filter: host_state = 0",
	}
	v1TableFilters["service_stale"] = []string{
		"Filter: service_staleness >= 1.5",
		"Filter: host_scheduled_downtime_depth = 0",
		"Filter: service_scheduled_downtime_depth = 0",
	}
}

type v1TableGetParams struct {
	Table  string    `path:"name"`
	Column *[]string `query:"column" description:"Columns to return" validate:"omitempty"`
	Filter *[]string `query:"filter" description:"Filter to apply on the table" validate:"omitempty"`
	Limit  *float64  `query:"limit" description:"Limit number of results" validate:"omitempty,min=0"`
}

func v1TableGet(c *gin.Context, params *v1TableGetParams) ([]gin.H, error) {
	client, err := GinGetLqlClient(c)
	if err != nil {
		return nil, err
	}
	user := c.GetString("user")

	columns := ""
	containsAll := false
	if params.Column != nil {
		for _, col := range *params.Column {
			if col == "all" {
				containsAll = true
				break
			}
		}
		columns = strings.Join(*params.Column, " ")
	} else if defaultCols, ok := v1TableColumns[params.Table]; ok {
		columns = strings.Join(defaultCols, " ")
	} else {
		columns = "name"
	}

	limit := 0
	if params.Limit != nil {
		limit = int(*params.Limit)
	}

	lines := []string{fmt.Sprintf("GET %s", params.Table)}
	if !containsAll {
		lines = append(lines, fmt.Sprintf("Columns: %s", columns))
	}

	if params.Filter != nil {
		filters := []string{}
		for _, filter := range *params.Filter {
			if addFilters, ok := v1TableFilters[filter]; ok {
				filters = append(filters, addFilters...)
				continue
			}
			if filter[0:7] != "Filter:" {
				return nil, fmt.Errorf("Invalid Filter '%s' given", filter)
			}

			filters = append(filters, filter)
		}
		lines = append(lines, filters...)
	}

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

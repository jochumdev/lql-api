package lql

import (
	"github.com/loopfz/gadgeto/tonic"
	"github.com/wI2L/fizz"
)

func v1Routes(grp *fizz.RouterGroup) {
	grp.POST("/raw", []fizz.OperationOption{
		fizz.Summary("GET RAW LQL Data or execute a COMMAND"),
		fizz.Response("400", "Bad request", nil, nil),
	}, tonic.Handler(v1RawPost, 200))

	grp.GET("/stats/tactical_overview", []fizz.OperationOption{
		fizz.Summary("GET tactical overview data"),
		fizz.Response("400", "Bad request", nil, nil),
	}, tonic.Handler(v1StatsGetTacticalOverview, 200))
}

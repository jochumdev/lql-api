package lql

import (
	"github.com/loopfz/gadgeto/tonic"
	"github.com/wI2L/fizz"
)

func v1Routes(grp *fizz.RouterGroup) {
	// // Add a new fruit to the market.
	// grp.POST("", []fizz.OperationOption{
	// 	fizz.Summary("Add a fruit to the market"),
	// 	fizz.Response("400", "Bad request", nil, nil),
	// }, tonic.Handler(CreateFruit, 200))

	// // Remove a fruit from the market,
	// // probably because it rotted.
	// grp.DELETE("/:name", []fizz.OperationOption{
	// 	fizz.Summary("Remove a fruit from the market"),
	// 	fizz.Response("400", "Fruit not found", nil, nil),
	// }, tonic.Handler(DeleteFruit, 204))

	// // List all available fruits.
	grp.POST("/raw", []fizz.OperationOption{
		fizz.Summary("GET RAW LQL Data or execute a COMMAND"),
		fizz.Response("400", "Bad request", nil, nil),
	}, tonic.Handler(v1RawPost, 200))
}

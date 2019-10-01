package restapi

import (
	"context"
	fcapi "github.com/carbonfive/go-filecoin-rest-api"

	"github.com/filecoin-project/go-filecoin/actor"
	"github.com/filecoin-project/go-filecoin/address"
)

// RESTPorcelain is the subset of porcelain and plumbing commands needed for the
// REST API callbacks
type RESTPorcelain interface {
	ActorGet(context.Context, address.Address) (*actor.Actor, error)
	ConfigGet(string) (interface{}, error)
}

// Launch creates and launches the REST API, serving HTTP from the given
// port
func Launch(ctx context.Context, porc RESTPorcelain, port int) *fcapi.HTTPAPI {
	api := fcapi.NewHTTPAPI(ctx, NewV1Callbacks(ctx, porc), port)
	api.Run()
	return api
}

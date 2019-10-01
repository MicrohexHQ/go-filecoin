package restapi

import (
	"context"

	fcapi "github.com/carbonfive/go-filecoin-rest-api"

	"github.com/filecoin-project/go-filecoin/address"
)

// NewV1Callbacks creates a facapi.V1Callbacks struct with callbacks
// that our porcelain calls
func NewV1Callbacks(ctx context.Context, porc RESTPorcelain) *fcapi.V1Callbacks {
	return &fcapi.V1Callbacks{
		Actor: MakeActorCallback(ctx, porc),
		Node:  MakeNodeCallback(porc),
	}
}

// MakeActorCallback makes an Actor callback function
func MakeActorCallback(ctx context.Context, porc RESTPorcelain) func(string) ([]byte, error) {
	return func(actorId string) (json []byte, err error) {
		addr, err := address.NewFromString(actorId)
		if err != nil {
			return []byte{}, err
		}
		actor, err := porc.ActorGet(ctx, addr)
		if err != nil {
			return []byte{}, err
		}
		return actor.Marshal()
	}
}

// MakeNodeCallback makes a Node callback function
func MakeNodeCallback(porc RESTPorcelain) func() ([]byte, error) {
	return func() (json []byte, err error) {
		ret, err := porc.ConfigGet("wallet.defaultAddress")
		addr := ret.(address.Address)
		if err != nil {
			undef, _ := address.Undef.MarshalJSON()
			return undef, err
		}

		return addr.MarshalJSON()
	}
}

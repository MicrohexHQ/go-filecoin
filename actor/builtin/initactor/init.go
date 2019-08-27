package initactor

import (
	"github.com/filecoin-project/go-filecoin/types"
	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"

	"github.com/filecoin-project/go-filecoin/abi"
	"github.com/filecoin-project/go-filecoin/actor"
	"github.com/filecoin-project/go-filecoin/exec"
	"github.com/filecoin-project/go-filecoin/vm/errors"
)

func init() {
	cbor.RegisterCborType(State{})
}

// Actor is the builtin actor responsible for network initialization.
// More details on future responsibilities can be found at https://github.com/filecoin-project/specs/blob/master/actors.md#init-actor.
type Actor struct{}

// State is the init actor's storage.
type State struct {
	Network string
}

// Ensure InitActor is an ExecutableActor at compile time.
var _ exec.ExecutableActor = (*Actor)(nil)

// initExports are the publicly (externally callable) methods of the AccountActor.
var initExports = exec.Exports{
	"getNetwork": &exec.FunctionSignature{
		Params: []abi.Type{},
		Return: []abi.Type{abi.String},
	},
}

// Exports makes the available methods for this contract available.
func (a *Actor) Exports() exec.Exports {
	return initExports
}

// NewActor returns a new storage market actor.
func NewActor() *actor.Actor {
	return actor.NewActor(types.InitActorCodeCid, types.ZeroAttoFIL)
}

// InitializeState for init actor.
func (ia *Actor) InitializeState(storage exec.Storage, networkInterface interface{}) error {
	network := networkInterface.(string)

	initStorage := &State{
		Network: network,
	}
	stateBytes, err := cbor.DumpObject(initStorage)
	if err != nil {
		return err
	}

	id, err := storage.Put(stateBytes)
	if err != nil {
		return err
	}

	return storage.Commit(id, cid.Undef)
}

// GetNetwork returns the network name for this network
func (sma *Actor) GetNetwork(vmctx exec.VMContext) (string, uint8, error) {
	if err := vmctx.Charge(actor.DefaultGasCost); err != nil {
		return "", exec.ErrInsufficientGas, errors.RevertErrorWrap(err, "Insufficient gas")
	}

	var state State
	err := actor.ReadState(vmctx, &state)
	if err != nil {
		return "", errors.CodeError(err), err
	}

	return state.Network, 0, nil
}

package restapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/go-filecoin/actor"
	"github.com/filecoin-project/go-filecoin/address"
	. "github.com/filecoin-project/go-filecoin/rest-api"
	"github.com/filecoin-project/go-filecoin/testhelpers"
)

func TestLaunchHappyPath(t *testing.T) {
	tc := requireTestCID(t, []byte("nothing"))
	actor1 := actor.Actor{Code: tc}
	defaultAddr := address.TestAddress

	porc := TestPorcelain{actor: &actor1, walletAddr: defaultAddr}

	port, err := testhelpers.GetFreePort()
	require.NoError(t, err)
	api := Launch(context.Background(), &porc, port)
	defer func() {
		err := api.Shutdown()
		if err != nil {
			t.Log(err)
		}
	}()

	t.Run("actor endpoint returns actor", func(t *testing.T) {
		path := fmt.Sprintf("actors/%s", defaultAddr.String())
		resp := RequireGetResponseBody(t, port, path)
		var actor2 actor.Actor
		assert.NoError(t, actor.UnmarshalStorage(resp, &actor2))
		assert.True(t, actor2.Code.Equals(actor1.Code))
	})

	t.Run("node endpoint returns correct value", func(t *testing.T) {
		resp := RequireGetResponseBody(t, port, "control/node")
		var node string
		require.NoError(t, json.Unmarshal(resp, &node))
		assert.Equal(t, node, defaultAddr.String())
	})
}

type TestPorcelain struct {
	walletAddr                  address.Address
	actor                       *actor.Actor
	failActorGet, failConfigGet bool
}

func (tp *TestPorcelain) ActorGet(ctx context.Context, addr address.Address) (*actor.Actor, error) {
	if tp.failActorGet {
		return nil, errors.New("ActorGet failed")
	}
	return tp.actor, nil
}

func (tp *TestPorcelain) ConfigGet(config string) (interface{}, error) {
	if tp.failConfigGet {
		return "", errors.New("ConfigGet failed")
	}
	if config == "wallet.defaultAddress" {
		return tp.walletAddr, nil
	}
	return "", errors.New("bad config call")
}

func RequireGetResponseBody(t *testing.T, port int, path string) []byte {
	uri := fmt.Sprintf("http://localhost:%d/%s", port, path)
	resp, err := http.Get(uri)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	return body
}

func requireTestCID(t *testing.T, data []byte) cid.Cid {
	hash, err := multihash.Sum(data, multihash.SHA2_256, -1)
	require.NoError(t, err)
	return cid.NewCidV1(cid.DagCBOR, hash)
}

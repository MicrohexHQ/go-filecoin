package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/go-filecoin/fixtures"
	th "github.com/filecoin-project/go-filecoin/testhelpers"
	tf "github.com/filecoin-project/go-filecoin/testhelpers/testflags"
	"github.com/filecoin-project/go-filecoin/tools/fast"
	"github.com/filecoin-project/go-filecoin/tools/fast/fastesting"
	"github.com/filecoin-project/go-filecoin/tools/fast/series"
	"github.com/filecoin-project/go-filecoin/types"
)

func TestChainHead(t *testing.T) {
	tf.IntegrationTest(t)

	d := th.NewDaemon(t).Start()
	defer d.ShutdownSuccess()

	jsonResult := d.RunSuccess("chain", "head", "--enc", "json").ReadStdoutTrimNewlines()

	var cidsFromJSON []cid.Cid
	err := json.Unmarshal([]byte(jsonResult), &cidsFromJSON)
	assert.NoError(t, err)

	textResult := d.RunSuccess("chain", "ls", "--enc", "text").ReadStdoutTrimNewlines()

	textCid, err := cid.Decode(textResult)
	require.NoError(t, err)

	assert.Equal(t, textCid, cidsFromJSON[0])
}

func TestChainLs(t *testing.T) {
	tf.IntegrationTest(t)

	t.Run("chain ls with json encoding returns the whole chain as json", func(t *testing.T) {
		d := makeTestDaemonWithMinerAndStart(t)
		defer d.ShutdownSuccess()

		op1 := d.RunSuccess("mining", "once", "--enc", "text")
		result1 := op1.ReadStdoutTrimNewlines()
		c, err := cid.Parse(result1)
		require.NoError(t, err)

		op2 := d.RunSuccess("chain", "ls", "--enc", "json")
		result2 := op2.ReadStdoutTrimNewlines()

		var bs [][]types.Block
		for _, line := range bytes.Split([]byte(result2), []byte{'\n'}) {
			var b []types.Block
			err := json.Unmarshal(line, &b)
			require.NoError(t, err)
			bs = append(bs, b)
			require.Equal(t, 1, len(b))
			line = bytes.TrimPrefix(line, []byte{'['})
			line = bytes.TrimSuffix(line, []byte{']'})

			// ensure conformance with JSON schema
			requireSchemaConformance(t, line, "filecoin_block")
		}

		assert.Equal(t, 2, len(bs))
		assert.True(t, bs[1][0].Parents.Empty())
		assert.True(t, c.Equals(bs[0][0].Cid()))
	})

	t.Run("chain ls with chain of size 1 returns genesis block", func(t *testing.T) {
		d := th.NewDaemon(t).Start()
		defer d.ShutdownSuccess()

		op := d.RunSuccess("chain", "ls", "--enc", "json")
		result := op.ReadStdoutTrimNewlines()

		var b []types.Block
		err := json.Unmarshal([]byte(result), &b)
		require.NoError(t, err)

		assert.True(t, b[0].Parents.Empty())
	})

	t.Run("chain ls with text encoding returns only CIDs", func(t *testing.T) {
		daemon := makeTestDaemonWithMinerAndStart(t)
		defer daemon.ShutdownSuccess()

		var blocks []types.Block
		blockJSON := daemon.RunSuccess("chain", "ls", "--enc", "json").ReadStdoutTrimNewlines()
		err := json.Unmarshal([]byte(blockJSON), &blocks)
		genesisBlockCid := blocks[0].Cid().String()
		require.NoError(t, err)

		newBlockCid := daemon.RunSuccess("mining", "once", "--enc", "text").ReadStdoutTrimNewlines()

		expectedOutput := fmt.Sprintf("%s\n%s", newBlockCid, genesisBlockCid)

		chainLsResult := daemon.RunSuccess("chain", "ls").ReadStdoutTrimNewlines()

		assert.Equal(t, chainLsResult, expectedOutput)
	})

	t.Run("chain ls --long returns CIDs, Miner, block height and message count", func(t *testing.T) {
		daemon := makeTestDaemonWithMinerAndStart(t)
		defer daemon.ShutdownSuccess()

		newBlockCid := daemon.RunSuccess("mining", "once", "--enc", "text").ReadStdoutTrimNewlines()

		chainLsResult := daemon.RunSuccess("chain", "ls", "--long").ReadStdoutTrimNewlines()

		assert.Contains(t, chainLsResult, newBlockCid)
		assert.Contains(t, chainLsResult, fixtures.TestMiners[0])
		assert.Contains(t, chainLsResult, "1")
		assert.Contains(t, chainLsResult, "0")
	})

	t.Run("chain ls --long with JSON encoding returns integer string block height", func(t *testing.T) {
		daemon := makeTestDaemonWithMinerAndStart(t)
		defer daemon.ShutdownSuccess()

		daemon.RunSuccess("mining", "once", "--enc", "text")
		chainLsResult := daemon.RunSuccess("chain", "ls", "--long", "--enc", "json").ReadStdoutTrimNewlines()
		assert.Contains(t, chainLsResult, `"height":"0"`)
		assert.Contains(t, chainLsResult, `"height":"1"`)
	})
}

func TestChainImportExport(t *testing.T) {
	tf.IntegrationTest(t)

	ctx, env := fastesting.NewTestEnvironment(context.Background(), t, fast.FilecoinOpts{
		DaemonOpts: []fast.ProcessDaemonOption{fast.POBlockTime(50 * time.Millisecond)},
	})
	// Teardown after test ends
	defer func() {
		err := env.Teardown(ctx)
		require.NoError(t, err)
	}()
	genesisNode := env.GenesisMiner

	// create a node that will export its chain
	exportNode := env.RequireNewNodeWithFunds(1000)
	require.NoError(t, series.Connect(ctx, genesisNode, exportNode))

	// create blocks with no messages
	series.CtxMiningOnce(ctx)

	// create block with 1 message
	require.NoError(t, series.SendFilecoinDefaults(ctx, genesisNode, exportNode, 10))
	series.CtxMiningOnce(ctx)

	// create block with many messages
	for i := 0; i < 20; i++ {
		require.NoError(t, series.SendFilecoinDefaults(ctx, genesisNode, exportNode, 10))
		require.NoError(t, series.SendFilecoinDefaults(ctx, exportNode, genesisNode, 10))
	}
	series.CtxMiningOnce(ctx)

	// create some more empty blocks
	for i := 0; i < 5; i++ {
		series.CtxMiningOnce(ctx)
	}

	// assert the genesisNode and export node are on the same head
	geneHead, err := genesisNode.ChainHead(ctx)
	require.NoError(t, err)
	exportHead, err := exportNode.ChainHead(ctx)
	require.NoError(t, err)
	require.Equal(t, geneHead, exportHead)

	// export node exports its chain
	exfi, err := ioutil.TempFile("", "fast_export_node.car")
	require.NoError(t, err)
	defer func() { _ = exfi.Close() }()
	require.NoError(t, exportNode.ChainExport(ctx, exfi.Name()))

	// create a new node unconnected to the rest to import chain
	importNode := env.RequireNewNodeStarted()
	imfi := files.NewReaderFile(exfi)
	defer func() { _ = imfi.Close() }()
	require.NoError(t, importNode.ChainImport(ctx, imfi))

	// imported chain head should equal exported chain head.
	importHead, err := importNode.ChainHead(ctx)
	require.NoError(t, err)
	assert.Equal(t, exportHead, importHead)

	// sanity check with an out of sync node
	dummyNode := env.RequireNewNodeStarted()
	dummyHead, err := dummyNode.ChainHead(ctx)
	require.NoError(t, err)
	assert.NotEqual(t, importHead, dummyHead)

}

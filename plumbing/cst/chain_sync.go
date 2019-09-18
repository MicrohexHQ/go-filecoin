package cst

import (
	"context"
	"io"

	"github.com/ipfs/go-ipfs-blockstore"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/filecoin-project/go-filecoin/chain"
	"github.com/filecoin-project/go-filecoin/types"
)

var logSync = logging.Logger("plumbing/syncer")

type chainSync interface {
	HandleNewTipSet(context.Context, *types.ChainInfo, bool) error
	Status() chain.Status
}

// ChainSyncProvider provides access to chain sync operations and their status.
type ChainSyncProvider struct {
	self peer.ID
	bs   blockstore.Blockstore
	sync chainSync
}

// NewChainSyncProvider returns a new ChainSyncProvider.
func NewChainSyncProvider(chainSyncer chainSync, bs blockstore.Blockstore, self peer.ID) *ChainSyncProvider {
	return &ChainSyncProvider{
		bs:   bs,
		self: self,
		sync: chainSyncer,
	}
}

// Status returns the chains current status, this includes whether or not the syncer is currently
// running, the chain being synced, and the time it started processing said chain.
func (chs *ChainSyncProvider) Status() chain.Status {
	return chs.sync.Status()
}

// HandleNewTipSet extends the Syncer's chain store with the given tipset if they
// represent a valid extension. It limits the length of new chains it will
// attempt to validate and caches invalid blocks it has encountered to
// help prevent DOS.
func (chs *ChainSyncProvider) HandleNewTipSet(ctx context.Context, ci *types.ChainInfo, trusted bool) error {
	return chs.sync.HandleNewTipSet(ctx, ci, trusted)
}

// ChainImport imports a chain with the same genesis block from `in`.
func (chs *ChainSyncProvider) ChainImport(ctx context.Context, in io.Reader) error {
	logSync.Info("starting CAR file import")
	headKey, err := chain.Import(ctx, chs.bs, in)
	if err != nil {
		return err
	}
	logSync.Infof("imported CAR file with head: %s", headKey)
	ci := &types.ChainInfo{
		Head: headKey,
		// TODO remove the below 2 lines when we can import chains without a fetcher.
		Peer:   chs.self,
		Height: 0,
	}
	logSync.Infof("processing car file with head: %s", headKey)
	if err := chs.HandleNewTipSet(ctx, ci, true); err != nil {
		logSync.Errorf("processing car file with head %s failed:%s", headKey, err)
		return err
	}
	logSync.Infof("processing car file with head: %s complete", headKey)
	return nil
}

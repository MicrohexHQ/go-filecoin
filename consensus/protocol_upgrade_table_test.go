package consensus

import (
	tf "github.com/filecoin-project/go-filecoin/testhelpers/testflags"
	"github.com/filecoin-project/go-filecoin/types"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestUpgradeTable(t *testing.T) {
	tf.UnitTest(t)

	t.Run("add single upgrade", func(t *testing.T) {
		network := "testnetwork"
		version := uint64(3)
		put := NewProtocolUpgradeTableBuilder(network).
			Add(network, version, types.NewBlockHeight(0)).
			Build()

		versionAtHeight, err := put.VersionAt(types.NewBlockHeight(0))
		require.NoError(t, err)

		assert.Equal(t, version, versionAtHeight)

		versionAtHeight, err = put.VersionAt(types.NewBlockHeight(1000))
		require.NoError(t, err)

		assert.Equal(t, version, versionAtHeight)
	})

	t.Run("finds correct version", func(t *testing.T) {
		network := "testnetwork"
		// add out of order and expect table to sort
		put := NewProtocolUpgradeTableBuilder(network).
			Add(network, 2, types.NewBlockHeight(20)).
			Add(network, 4, types.NewBlockHeight(40)).
			Add(network, 3, types.NewBlockHeight(30)).
			Add(network, 1, types.NewBlockHeight(10)).
			Add(network, 0, types.NewBlockHeight(0)).
			Build()

		for i := uint64(0); i < 50; i++ {
			version, err := put.VersionAt(types.NewBlockHeight(i))
			require.NoError(t, err)

			assert.Equal(t, i/10, version)
		}
	})

	t.Run("retrieving from empty table is an error", func(t *testing.T) {
		network := "testnetwork"
		put := NewProtocolUpgradeTableBuilder(network).Build()

		_, err := put.VersionAt(types.NewBlockHeight(0))
		require.Error(t, err)
		assert.Matches(t, err.Error(), "no protocol versions")
	})

	t.Run("retrieving before first upgrade is an error", func(t *testing.T) {
		network := "testnetwork"
		put := NewProtocolUpgradeTableBuilder(network).
			Add(network, 2, types.NewBlockHeight(20)).
			Build()

		_, err := put.VersionAt(types.NewBlockHeight(0))
		require.Error(t, err)
		assert.Matches(t, err.Error(), "less than effective start")
	})

	t.Run("ignores upgrades from wrong network", func(t *testing.T) {
		network := "testnetwork"
		otherNetwork := "othernetwork"

		put := NewProtocolUpgradeTableBuilder(network).
			Add(network, 0, types.NewBlockHeight(0)).
			Add(otherNetwork, 1, types.NewBlockHeight(10)).
			Add(otherNetwork, 2, types.NewBlockHeight(20)).
			Add(network, 3, types.NewBlockHeight(30)).
			Add(otherNetwork, 4, types.NewBlockHeight(40)).
			Build()

		for i := uint64(0); i < 50; i++ {
			version, err := put.VersionAt(types.NewBlockHeight(i))
			require.NoError(t, err)

			expectedVersion := uint64(0)
			if i >= 30 {
				expectedVersion = 3
			}
			assert.Equal(t, expectedVersion, version)
		}
	})

}

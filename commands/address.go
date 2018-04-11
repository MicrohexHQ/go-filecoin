package commands

import (
	"fmt"
	"io"

	cmds "gx/ipfs/QmYMj156vnPY7pYvtkvQiMDAzqWDDHkfiW5bYbMpYoHxhB/go-ipfs-cmds"
	cmdkit "gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit"

	"github.com/filecoin-project/go-filecoin/types"
)

var walletCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Manage your filecoin wallets",
	},
	Subcommands: map[string]*cmds.Command{
		"addrs":   addrsCmd,
		"balance": balanceCmd,
	},
}

var addrsCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Interact with addresses",
	},
	Subcommands: map[string]*cmds.Command{
		"ls":     addrsListCmd,
		"new":    addrsNewCmd,
		"lookup": addrsLookupCmd,
	},
}

type addressResult struct {
	Address string
}

var addrsNewCmd = &cmds.Command{
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		fcn := GetNode(env)
		re.Emit(&addressResult{fcn.Wallet.NewAddress().String()}) // nolint: errcheck
		return nil
	},
	Type: &addressResult{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, a *addressResult) error {
			_, err := fmt.Fprintln(w, a.Address)
			return err
		}),
	},
}

var addrsListCmd = &cmds.Command{
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		fcn := GetNode(env)
		for _, a := range fcn.Wallet.GetAddresses() {
			re.Emit(&addressResult{a.String()}) // nolint: errcheck
		}
		return nil
	},
	Type: &addressResult{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, addr *addressResult) error {
			_, err := fmt.Fprintln(w, addr.Address)
			return err
		}),
	},
}

var addrsLookupCmd = &cmds.Command{
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("address", true, false, "address to find peerId for"),
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		fcn := GetNode(env)

		address, err := types.NewAddressFromString(req.Arguments[0])
		if err != nil {
			return err
		}

		v, err := fcn.Lookup.Lookup(req.Context, address)
		if err != nil {
			return err
		}
		re.Emit(v.Pretty()) // nolint: errcheck
		return nil
	},
	Type: string(""),
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, pid string) error {
			_, err := fmt.Fprintln(w, pid)
			return err
		}),
	},
}
var balanceCmd = &cmds.Command{
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("address", true, false, "address to get balance for"),
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		fcn := GetNode(env)
		blk := fcn.ChainMgr.GetBestBlock()
		if blk.StateRoot == nil {
			return ErrLatestBlockStateRootNil
		}

		tree, err := types.LoadStateTree(req.Context, fcn.CborStore, blk.StateRoot)
		if err != nil {
			return err
		}

		addr, err := types.NewAddressFromString(req.Arguments[0])
		if err != nil {
			return err
		}

		act, err := tree.GetActor(req.Context, addr)
		if err != nil {
			return err
		}

		re.Emit(act.Balance) // nolint: errcheck
		return nil
	},
	Type: &types.TokenAmount{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, b *types.TokenAmount) error {
			return PrintString(w, b)
		}),
	},
}

// Package commands implements the command to print the blockchain.
package commands

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-cmdkit"
	"github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs-files"

	"github.com/filecoin-project/go-filecoin/types"
)

var chainCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Inspect the filecoin blockchain",
	},
	Subcommands: map[string]*cmds.Command{
		"export": storeExportCmd,
		"head":   storeHeadCmd,
		"import": storeImportCmd,
		"ls":     storeLsCmd,
		"status": storeStatusCmd,
	},
}

var storeHeadCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Get heaviest tipset CIDs",
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		head, err := GetPorcelainAPI(env).ChainHead()
		if err != nil {
			return err
		}
		return re.Emit(head.Key())
	},
	Type: []cid.Cid{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, res []cid.Cid) error {
			for _, r := range res {
				_, err := fmt.Fprintln(w, r.String())
				if err != nil {
					return err
				}
			}
			return nil
		}),
	},
}

var storeLsCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline:          "List blocks in the blockchain",
		ShortDescription: `Provides a list of blocks in order from head to genesis. By default, only CIDs are returned for each block.`,
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption("long", "l", "List blocks in long format, including CID, Miner, StateRoot, block height and message count respectively"),
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		iter, err := GetPorcelainAPI(env).ChainLs(req.Context)
		if err != nil {
			return err
		}
		for ; !iter.Complete(); err = iter.Next() {
			if err != nil {
				return err
			}
			if !iter.Value().Defined() {
				panic("tipsets from this iterator should have at least one member")
			}
			if err := re.Emit(iter.Value().ToSlice()); err != nil {
				return err
			}
		}
		return nil
	},
	Type: []types.Block{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, res *[]types.Block) error {
			showAll, _ := req.Options["long"].(bool)
			blocks := *res

			for _, block := range blocks {
				var output strings.Builder

				if showAll {
					output.WriteString(block.Cid().String())
					output.WriteString("\t")
					output.WriteString(block.Miner.String())
					output.WriteString("\t")
					output.WriteString(block.StateRoot.String())
					output.WriteString("\t")
					output.WriteString(strconv.FormatUint(uint64(block.Height), 10))
					output.WriteString("\t")
					output.WriteString(block.Messages.String())
				} else {
					output.WriteString(block.Cid().String())
				}

				_, err := fmt.Fprintln(w, output.String())
				if err != nil {
					return err
				}
			}

			return nil
		}),
	},
}

var storeStatusCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Show status of chain sync operation.",
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		syncStatus := GetPorcelainAPI(env).ChainStatus()
		if err := re.Emit(syncStatus); err != nil {
			return err
		}
		return nil
	},
}

var storeExportCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Export the chain store to a car file.",
	},
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("file", true, false, "File to export chain data to."),
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		f, err := os.Create(req.Arguments[0])
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		return GetPorcelainAPI(env).ChainExport(req.Context, f)
	},
}

var storeImportCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Import the chain from a car file.",
	},
	Arguments: []cmdkit.Argument{
		cmdkit.FileArg("file", true, false, "File to import chain data from.").EnableStdin(),
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		iter := req.Files.Entries()
		if !iter.Next() {
			return fmt.Errorf("no file given: %s", iter.Err())
		}

		fi, ok := iter.Node().(files.File)
		if !ok {
			return fmt.Errorf("given file was not a files.File")
		}
		defer func() { _ = fi.Close() }()
		return GetPorcelainAPI(env).ChainImport(req.Context, fi)
	},
}

package fast

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-files"

	"github.com/filecoin-project/go-filecoin/chain"
)

// ChainHead runs the chain head command against the filecoin process.
func (f *Filecoin) ChainHead(ctx context.Context) ([]cid.Cid, error) {
	var out []cid.Cid
	if err := f.RunCmdJSONWithStdin(ctx, nil, &out, "go-filecoin", "chain", "head"); err != nil {
		return nil, err
	}
	return out, nil

}

// ChainLs runs the chain ls command against the filecoin process.
func (f *Filecoin) ChainLs(ctx context.Context) (*json.Decoder, error) {
	return f.RunCmdLDJSONWithStdin(ctx, nil, "go-filecoin", "chain", "ls")
}

// ChainStatus runs the chain status command against the filecoin process.
func (f *Filecoin) ChainStatus(ctx context.Context) (*chain.Status, error) {
	var out *chain.Status
	if err := f.RunCmdJSONWithStdin(ctx, nil, &out, "go-filecoin", "chain", "status"); err != nil {
		return nil, err
	}
	return out, nil
}

// ChainExport runs the chain export command against the filecoin process.
func (f *Filecoin) ChainExport(ctx context.Context, filepath string) error {
	out, err := f.RunCmdWithStdin(ctx, nil, "go-filecoin", "chain", "export", filepath)
	if err != nil {
		return err
	}
	if out.ExitCode() > 0 {
		return fmt.Errorf("filecoin command: %s, exited with non-zero exitcode: %d", out.Args(), out.ExitCode())
	}
	return nil
}

// ChainImport runs the chain import command against the filecoin process.
func (f *Filecoin) ChainImport(ctx context.Context, file files.File) error {
	out, err := f.RunCmdWithStdin(ctx, file, "go-filecoin", "chain", "import")
	if err != nil {
		return err
	}
	if out.ExitCode() > 0 {
		return fmt.Errorf("filecoin command: %s, exited with non-zero exitcode: %d", out.Args(), out.ExitCode())
	}
	return nil
}

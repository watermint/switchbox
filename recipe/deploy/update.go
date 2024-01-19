package deploy

import (
	"github.com/watermint/switchbox/domain/sb_deploy"
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_auth"
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_conn"
	"github.com/watermint/toolbox/essentials/model/mo_path"
	"github.com/watermint/toolbox/essentials/model/mo_string"
	"github.com/watermint/toolbox/infra/control/app_control"
	"github.com/watermint/toolbox/quality/infra/qt_errors"
)

type Update struct {
	Peer           dbx_conn.ConnScopedIndividual
	SourceUrl      string
	SourcePassword mo_string.OptionalString
	Prefix         string
	Suffix         string
	CellarPath     mo_path.FileSystemPath
	BinaryName     string
}

func (z *Update) Preset() {
	z.Peer.SetScopes(
		dbx_auth.ScopeFilesContentRead,
		dbx_auth.ScopeFilesMetadataRead,
		dbx_auth.ScopeSharingRead,
	)
}

func (z *Update) Exec(c app_control.Control) error {
	recipe := sb_deploy.BinSrcDropboxDstLocal{
		SourceUrl:      z.SourceUrl,
		SourcePassword: z.SourcePassword.Value(),
		BinaryName:     z.BinaryName,
		Prefix:         z.Prefix,
		Suffix:         z.Suffix,
		CellarPath:     z.CellarPath.Path(),
	}

	worker := sb_deploy.NewBinSrcDropboxDstLocal(recipe, c, z.Peer.Client())
	if err := worker.UpdateIfRequired(); err != nil {
		return err
	}
	return nil
}

func (z *Update) Test(c app_control.Control) error {
	return qt_errors.ErrorHumanInteractionRequired
}

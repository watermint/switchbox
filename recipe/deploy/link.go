package deploy

import (
	"github.com/watermint/switchbox/domain/sb_deploy"
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_auth"
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_conn"
	"github.com/watermint/toolbox/essentials/terminal/es_window"
	"github.com/watermint/toolbox/infra/control/app_control"
	"github.com/watermint/toolbox/infra/data/da_json"
	"github.com/watermint/toolbox/quality/infra/qt_errors"
	"os"
)

type Link struct {
	Peer   dbx_conn.ConnScopedIndividual
	Deploy da_json.JsonInput
	Force  bool
	Hide   bool
}

func (z *Link) Preset() {
	z.Peer.SetScopes(
		dbx_auth.ScopeFilesContentRead,
		dbx_auth.ScopeFilesMetadataRead,
		dbx_auth.ScopeSharingRead,
	)
	z.Deploy.SetModel(&sb_deploy.BinSrcDropboxDstLocalRecipe{})
}

func (z *Link) Exec(c app_control.Control) error {
	l := c.Log()
	if z.Hide {
		es_window.HideConsole()
		l.Info("Hide console")
	}

	var deploy *sb_deploy.BinSrcDropboxDstLocalRecipe
	if v, err := z.Deploy.Unmarshal(); err != nil {
		return err
	} else {
		deploy = v.(*sb_deploy.BinSrcDropboxDstLocalRecipe)
	}

	worker := sb_deploy.NewBinSrcDropboxDstLocal(*deploy, c, z.Peer.Client())

	shouldUpdate := z.Force
	if !shouldUpdate {
		updateRequired, err := worker.IsUpdateRequired()
		if err != nil {
			return err
		}
		shouldUpdate = updateRequired
		if _, err := os.Lstat(deploy.DeployPath); os.IsNotExist(err) {
			shouldUpdate = true
		}
	}

	if err := worker.UpdateIfRequired(); err != nil {
		return err
	}

	if shouldUpdate {
		l.Info("Update required")
		return worker.DeploySymlink()
	} else {
		l.Info("No update required")
		return nil
	}
}

func (z *Link) Test(c app_control.Control) error {
	return qt_errors.ErrorHumanInteractionRequired
}

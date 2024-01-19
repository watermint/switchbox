package deploy

import (
	"github.com/watermint/switchbox/domain/sb_deploy"
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_auth"
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_conn"
	"github.com/watermint/toolbox/essentials/terminal/es_window"
	"github.com/watermint/toolbox/infra/control/app_control"
	"github.com/watermint/toolbox/infra/data/da_json"
	"github.com/watermint/toolbox/quality/infra/qt_errors"
)

type Update struct {
	Peer   dbx_conn.ConnScopedIndividual
	Deploy da_json.JsonInput
	Force  bool
	Hide   bool
}

func (z *Update) Preset() {
	z.Peer.SetScopes(
		dbx_auth.ScopeFilesContentRead,
		dbx_auth.ScopeFilesMetadataRead,
		dbx_auth.ScopeSharingRead,
	)
	z.Deploy.SetModel(&sb_deploy.BinSrcDropboxDstLocal{})
}

func (z *Update) Exec(c app_control.Control) error {
	l := c.Log()
	if z.Hide {
		es_window.HideConsole()
		l.Info("Hide console")
	}

	var deploy *sb_deploy.BinSrcDropboxDstLocal
	if v, err := z.Deploy.Unmarshal(); err != nil {
		return err
	} else {
		deploy = v.(*sb_deploy.BinSrcDropboxDstLocal)
	}

	worker := sb_deploy.NewBinSrcDropboxDstLocal(*deploy, c, z.Peer.Client())
	if z.Force {
		if err := worker.UpdateForce(); err != nil {
			return err
		}
	} else {
		if err := worker.UpdateIfRequired(); err != nil {
			return err
		}
	}
	return nil
}

func (z *Update) Test(c app_control.Control) error {
	return qt_errors.ErrorHumanInteractionRequired
}

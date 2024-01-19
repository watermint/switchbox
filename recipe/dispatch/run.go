package dispatch

import (
	"bufio"
	"encoding/json"
	"github.com/watermint/switchbox/domain/sb_deploy"
	"github.com/watermint/switchbox/domain/sb_dispatch"
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_auth"
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_conn"
	"github.com/watermint/toolbox/essentials/log/esl"
	"github.com/watermint/toolbox/infra/control/app_control"
	"github.com/watermint/toolbox/infra/data/da_json"
	"github.com/watermint/toolbox/quality/infra/qt_errors"
	"os"
	"os/exec"
)

type Run struct {
	Peer    dbx_conn.ConnScopedIndividual
	Runbook da_json.JsonInput
	Deploy  da_json.JsonInput
}

func (z *Run) Preset() {
	z.Peer.SetScopes(
		dbx_auth.ScopeFilesContentRead,
		dbx_auth.ScopeFilesMetadataRead,
		dbx_auth.ScopeSharingRead,
	)
	z.Runbook.SetModel(&sb_dispatch.BinRunbook{})
	z.Deploy.SetModel(&sb_deploy.BinSrcDropboxDstLocal{})
}

func (z *Run) Exec(c app_control.Control) error {
	l := c.Log()
	var runbook sb_dispatch.BinRunbook
	rbContent, err := os.ReadFile(z.Runbook.FilePath())
	if err != nil {
		l.Warn("Unable to read runbook", esl.Error(err))
		return err
	}
	if err := json.Unmarshal(rbContent, &runbook); err != nil {
		l.Warn("Unable to parse runbook", esl.Error(err))
		return err
	}
	var deploy sb_deploy.BinSrcDropboxDstLocal
	deployContent, err := os.ReadFile(z.Deploy.FilePath())
	if err != nil {
		l.Warn("Unable to read deploy", esl.Error(err))
		return err
	}
	if err := json.Unmarshal(deployContent, &deploy); err != nil {
		l.Warn("Unable to parse deploy", esl.Error(err))
		return err
	}

	deployWorker := sb_deploy.NewBinSrcDropboxDstLocal(deploy, c, z.Peer.Client())
	if err := deployWorker.UpdateIfRequired(); err != nil {
		return err
	}

	binPath := deployWorker.LocalLatestBinaryPath()
	if binPath == "" {
		c.Log().Warn("No binary found")
		return nil
	}
	l.Info("Run", esl.String("binPath", binPath), esl.Strings("args", runbook.Args))

	cmd := exec.Command(binPath, runbook.Args...)

	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		l.Warn("Unable to get stdout pipe", esl.Error(err))
		return err
	}
	cmdErr, err := cmd.StderrPipe()
	if err != nil {
		l.Warn("Unable to get stderr pipe", esl.Error(err))
		return err
	}
	scannerOut := bufio.NewScanner(cmdOut)
	scannerErr := bufio.NewScanner(cmdErr)
	go func() {
		for scannerOut.Scan() {
			l.Info("Out", esl.String("Line", scannerOut.Text()))
		}
	}()
	go func() {
		for scannerErr.Scan() {
			l.Info("Err", esl.String("Line", scannerErr.Text()))
		}
	}()
	if err := cmd.Start(); err != nil {
		l.Warn("Unable to start command", esl.Error(err))
		return err
	}
	if err := cmd.Wait(); err != nil {
		l.Warn("Command failed", esl.Error(err))
		return err
	}
	return nil
}

func (z *Run) Test(c app_control.Control) error {
	return qt_errors.ErrorHumanInteractionRequired
}

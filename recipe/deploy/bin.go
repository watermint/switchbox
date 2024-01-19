package deploy

import (
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_auth"
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_conn"
	"github.com/watermint/toolbox/domain/dropbox/model/mo_file"
	dbx_path "github.com/watermint/toolbox/domain/dropbox/model/mo_path"
	"github.com/watermint/toolbox/domain/dropbox/model/mo_url"
	"github.com/watermint/toolbox/domain/dropbox/service/sv_sharedlink_file"
	"github.com/watermint/toolbox/essentials/io/es_zip"
	"github.com/watermint/toolbox/essentials/strings/es_version"
	"github.com/watermint/toolbox/infra/control/app_definitions"

	"github.com/watermint/toolbox/essentials/log/esl"
	"github.com/watermint/toolbox/essentials/model/mo_path"
	"github.com/watermint/toolbox/essentials/model/mo_string"
	"github.com/watermint/toolbox/infra/control/app_control"
	"github.com/watermint/toolbox/quality/infra/qt_errors"
	"os"
	"path/filepath"
	"strings"
)

type Bin struct {
	Peer           dbx_conn.ConnScopedIndividual
	SourceUrl      string
	SourcePassword mo_string.OptionalString
	Prefix         string
	Suffix         string
	CellarPath     mo_path.FileSystemPath
	DeployPath     mo_path.FileSystemPath
	BinaryName     string
}

func (z *Bin) Preset() {
	z.Peer.SetScopes(
		dbx_auth.ScopeFilesContentRead,
		dbx_auth.ScopeFilesMetadataRead,
		dbx_auth.ScopeSharingRead,
	)
}

func (z *Bin) listLocalVersions(c app_control.Control) (versions []es_version.Version, versionPaths map[string]string, err error) {
	versions = make([]es_version.Version, 0)
	versionPaths = make(map[string]string)
	l := c.Log().With(esl.String("cellarPath", z.CellarPath.Path()))
	entries, err := os.ReadDir(z.CellarPath.Path())
	if err != nil {
		l.Debug("Unable to read cellar directory", esl.Error(err))
		return versions, versionPaths, err
	}
	fullPrefix := z.Prefix + "-"

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), fullPrefix) {
			verStr := strings.TrimPrefix(entry.Name(), fullPrefix)
			ver, err := es_version.Parse(verStr)
			if err != nil {
				l.Debug("Unable to parse version", esl.Error(err))
				continue
			}
			versions = append(versions, ver)
			versionPaths[ver.String()] = filepath.Join(z.CellarPath.Path(), entry.Name())
		}
	}

	return versions, versionPaths, nil
}

func (z *Bin) listRemoteVersions(c app_control.Control) (versions []es_version.Version, versionPaths map[string]string, err error) {
	l := c.Log().With(esl.String("sourceUrl", z.SourceUrl))
	versions = make([]es_version.Version, 0)
	versionPaths = make(map[string]string)

	url, err := mo_url.NewUrl(z.SourceUrl)
	if err != nil {
		l.Debug("Unable to parse url", esl.Error(err))
		return versions, versionPaths, err
	}
	fullPrefix := z.Prefix + "-"
	fullFileSuffix := "-" + z.Suffix + ".zip"

	svs := sv_sharedlink_file.New(z.Peer.Client())
	err = svs.List(url, dbx_path.NewDropboxPath(""), func(folderEntry mo_file.Entry) {
		if folder, ok := folderEntry.Folder(); !ok {
			l.Debug("Skip entry", esl.String("name", folderEntry.Name()))
		} else {
			folderPath := dbx_path.NewDropboxPath("").ChildPath(folder.Name())
			err = svs.List(url, folderPath, func(fileEntry mo_file.Entry) {
				if !strings.HasSuffix(fileEntry.Name(), fullFileSuffix) || !strings.HasPrefix(fileEntry.Name(), fullPrefix) {
					l.Debug("Skip entry", esl.String("name", fileEntry.Name()))
					return
				}
				verStr := strings.TrimPrefix(folder.Name(), fullPrefix)
				ver, err := es_version.Parse(verStr)
				if err != nil {
					l.Debug("Unable to parse version", esl.Error(err))
					return
				}
				if file, ok := fileEntry.File(); ok {
					l.Debug("Found version", esl.String("version", ver.String()), esl.String("path", file.Path().Path()))
					versions = append(versions, ver)
					versionPaths[ver.String()] = folderPath.ChildPath(file.Name()).Path()
				}
			})
		}
	})
	if err != nil {
		l.Debug("Unable to list remote versions", esl.Error(err))
		return versions, versionPaths, err
	}

	return versions, versionPaths, nil
}

func (z *Bin) downloadLatest(c app_control.Control, version es_version.Version, versionPath string) (downloadPath string, err error) {
	l := c.Log().With(esl.String("version", version.String()), esl.String("versionPath", versionPath))
	l.Debug("Download version")

	svs := sv_sharedlink_file.New(z.Peer.Client())
	url, err := mo_url.NewUrl(z.SourceUrl)
	if err != nil {
		l.Debug("Unable to parse url", esl.Error(err))
		return "", err
	}
	entry, path, err := svs.Download(url, dbx_path.NewDropboxPath(versionPath), z.DeployPath, sv_sharedlink_file.Password(z.SourcePassword.Value()))
	if err != nil {
		l.Debug("Unable to download version", esl.Error(err))
		return "", err
	}
	l.Info("Downloaded", esl.String("path", path.Path()), esl.String("entry", entry.Path().Path()))

	return path.Path(), nil
}

func (z *Bin) extractLatest(c app_control.Control, version es_version.Version, downloadPath string) (cellarPath string, err error) {
	l := c.Log().With(esl.String("downloadPath", downloadPath))
	l.Debug("Extract version")

	cellarPath = filepath.Join(z.CellarPath.Path(), z.Prefix+"-"+version.String())
	if err := os.MkdirAll(cellarPath, 0755); err != nil {
		l.Debug("Unable to create destination directory", esl.Error(err))
		return "", err
	}
	l.Info("Extracting into cellar directory", esl.String("cellarPath", cellarPath))
	err = es_zip.Extract(l, downloadPath, cellarPath)
	l.Debug("Extracted", esl.Error(err), esl.String("cellarPath", cellarPath))

	if err := os.Remove(downloadPath); err != nil {
		l.Warn("Unable to remove downloaded file", esl.Error(err))
	}

	return cellarPath, err
}

func (z *Bin) deploySymlink(c app_control.Control, version es_version.Version, cellarPath string) error {
	l := c.Log().With(esl.String("cellarPath", cellarPath))
	l.Info("Deploying symlink")

	var binName string
	if app_definitions.IsWindows() {
		binName = z.BinaryName + ".exe"
	} else {
		binName = z.BinaryName
	}
	binCellarPath := filepath.Join(z.CellarPath.Path(), z.Prefix+"-"+version.String(), binName)
	binDeployPath := filepath.Join(z.DeployPath.Path(), binName)

	_, err := os.Lstat(binDeployPath)
	if err != nil && !os.IsNotExist(err) {
		l.Warn("Unable to stat existing symlink", esl.Error(err))
		return err
	} else if err == nil || os.IsExist(err) {
		l.Info("Existing symlink found, try removing existing link", esl.String("binDeployPath", binDeployPath))
		if err := os.Remove(binDeployPath); err != nil {
			l.Warn("Unable to remove existing symlink", esl.Error(err))
			return err
		}
		l.Info("Existing symlink removed", esl.String("binDeployPath", binDeployPath))
	}
	if err := os.MkdirAll(z.DeployPath.Path(), 0755); err != nil {
		l.Warn("Unable to create deploy directory", esl.Error(err))
		return err
	}

	if err := os.Chmod(binCellarPath, 0755); err != nil {
		l.Warn("Unable to change permission", esl.Error(err))
		return err
	}

	if err := os.Symlink(binCellarPath, binDeployPath); err != nil {
		l.Warn("Unable to create symlink", esl.Error(err))
		return err
	}
	l.Info("Deployed", esl.String("binDeployPath", binDeployPath))
	return nil
}

func (z *Bin) Exec(c app_control.Control) error {
	l := c.Log()

	if err := os.MkdirAll(z.CellarPath.Path(), 0755); err != nil {
		l.Debug("Unable to create cellar directory", esl.Error(err))
		return err
	}

	localVersions, localVersionPaths, err := z.listLocalVersions(c)
	if err != nil {
		return err
	}
	remoteVersions, remoteVersionPaths, err := z.listRemoteVersions(c)
	if err != nil {
		return err
	}

	localVersionLatest := es_version.Max(localVersions...)
	localVersionLatestPath := localVersionPaths[localVersionLatest.String()]
	remoteVersionLatest := es_version.Max(remoteVersions...)
	remoteVersionLatestPath := remoteVersionPaths[remoteVersionLatest.String()]

	if localVersionLatest.Equals(remoteVersionLatest) {
		l.Info("Already latest version", esl.String("localVersion", localVersionLatest.String()),
			esl.String("localPath", localVersionLatestPath),
			esl.String("remoteVersion", remoteVersionLatest.String()),
			esl.String("remotePath", remoteVersionLatestPath))

		return nil
	}

	l.Info("Local latest version", esl.String("version", localVersionLatest.String()), esl.String("path", localVersionLatestPath))
	l.Info("Remote latest version", esl.String("version", remoteVersionLatest.String()), esl.String("path", remoteVersionLatestPath))

	dlPath, err := z.downloadLatest(c, remoteVersionLatest, remoteVersionLatestPath)
	if err != nil {
		return err
	}
	cellarPath, err := z.extractLatest(c, remoteVersionLatest, dlPath)
	if err != nil {
		return err
	}
	l.Info("Extracted", esl.String("path", cellarPath))

	return z.deploySymlink(c, remoteVersionLatest, cellarPath)
}

func (z *Bin) Test(c app_control.Control) error {
	return qt_errors.ErrorHumanInteractionRequired
}

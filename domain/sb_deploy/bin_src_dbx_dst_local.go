package sb_deploy

import (
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_client"
	"github.com/watermint/toolbox/domain/dropbox/model/mo_file"
	dbx_path "github.com/watermint/toolbox/domain/dropbox/model/mo_path"
	"github.com/watermint/toolbox/domain/dropbox/model/mo_url"
	"github.com/watermint/toolbox/domain/dropbox/service/sv_sharedlink_file"
	"github.com/watermint/toolbox/essentials/io/es_zip"
	"github.com/watermint/toolbox/essentials/log/esl"
	"github.com/watermint/toolbox/essentials/model/mo_path"
	"github.com/watermint/toolbox/essentials/strings/es_version"
	"github.com/watermint/toolbox/infra/control/app_control"
	"github.com/watermint/toolbox/infra/control/app_definitions"
	"os"
	"path/filepath"
	"strings"
)

// BinSrcDropboxDstLocal Deploy binary from Dropbox to local
// This recipe expect folder structure like this:
// `PREFIX-VERSION/PREFIX-VERSION-SUFFIX.zip`.
// For example, if prefix is `myapp` and suffix is `linux-amd64`, the folder structure is:
// `myapp-1.0.0/myapp-1.0.0-linux-amd64.zip`.
type BinSrcDropboxDstLocal struct {
	// SourceUrl is the url to the shared link folder
	SourceUrl string `json:"source_url"`

	// SourcePassword is the password to access the source. If empty, no password is used.
	SourcePassword string `json:"source_password,omitempty"`

	// BinaryName is the name of the binary file
	BinaryName string `json:"binary_name"`

	// Prefix is the prefix of the file/folder basename.
	Prefix string `json:"prefix"`

	// Suffix is the suffix of the file basename.
	Suffix string `json:"suffix"`

	// CellarPath is the path to store extracted binaries of versions
	CellarPath string `json:"cellar_path"`

	// DeployPath is the path to deploy symlink to the binary.
	// This field is options when no symlink deployment is required.
	DeployPath string `json:"deploy_path,omitempty"`
}

type BinSrcDropboxDstLocalWorker interface {
	// UpdateIfRequired Update if required
	UpdateIfRequired() (err error)

	// GetLocalLatest Get local latest version and download automatically if
	// no local version found. But this will not check remote versions if local
	// version is found.
	GetLocalLatest() (binaryPath string, version es_version.Version, err error)

	// ListLocalVersions List local versions
	ListLocalVersions() (versions []es_version.Version, versionPaths map[string]string, err error)

	// ListRemoteVersions List remote versions
	ListRemoteVersions() (versions []es_version.Version, versionPaths map[string]string, err error)

	// Download version to temporary path
	Download(version es_version.Version, versionPath string) (downloadPath string, err error)

	// Extract version into cellar
	Extract(version es_version.Version, downloadPath string) (cellarPath string, err error)

	// DeploySymlink Deploy latest binary as symlink
	DeploySymlink() (err error)

	// BinaryName returns the binary name that consider current OS platform
	BinaryName() string

	// LocalLatestBinaryPath returns the path to the latest binary. Returns empty string
	// if no local version found.
	LocalLatestBinaryPath() string
}

func NewBinSrcDropboxDstLocal(recipe BinSrcDropboxDstLocal, ctl app_control.Control, client dbx_client.Client) BinSrcDropboxDstLocalWorker {
	return &binSrcDropboxDstLocalWorkerImpl{
		recipe: recipe,
		ctl:    ctl,
		client: client,
	}
}

type binSrcDropboxDstLocalWorkerImpl struct {
	recipe BinSrcDropboxDstLocal
	ctl    app_control.Control
	client dbx_client.Client
}

func (z binSrcDropboxDstLocalWorkerImpl) LocalLatestBinaryPath() string {
	l := z.ctl.Log()
	localVersions, localVersionPaths, err := z.ListLocalVersions()
	if err != nil {
		l.Debug("Unable to list local versions", esl.Error(err))
		return ""
	}
	localVersionLatest := es_version.Max(localVersions...)
	if len(localVersions) < 1 || localVersionLatest.Equals(es_version.Zero()) {
		return ""
	}
	return filepath.Join(localVersionPaths[localVersionLatest.String()], z.BinaryName())
}

func (z binSrcDropboxDstLocalWorkerImpl) BinaryName() string {
	if app_definitions.IsWindows() {
		return z.recipe.BinaryName + ".exe"
	} else {
		return z.recipe.BinaryName
	}
}

func (z binSrcDropboxDstLocalWorkerImpl) UpdateIfRequired() (err error) {
	l := z.ctl.Log()

	if err := os.MkdirAll(z.recipe.CellarPath, 0755); err != nil {
		l.Warn("Unable to create cellar directory", esl.Error(err))
		return err
	}

	localVersions, localVersionPaths, err := z.ListLocalVersions()
	if err != nil {
		l.Warn("Unable to list local versions", esl.Error(err))
		return err
	}
	remoteVersions, remoteVersionPaths, err := z.ListRemoteVersions()
	if err != nil {
		l.Warn("Unable to list remote versions", esl.Error(err))
		return err
	}

	localVersionLatest := es_version.Max(localVersions...)
	remoteVersionLatest := es_version.Max(remoteVersions...)

	if localVersionLatest.Equals(remoteVersionLatest) {
		l.Info("Already latest version", esl.String("localVersion", localVersionLatest.String()),
			esl.String("localPath", localVersionPaths[localVersionLatest.String()]),
			esl.String("remoteVersion", remoteVersionLatest.String()),
			esl.String("remotePath", remoteVersionPaths[remoteVersionLatest.String()]))
		return nil
	}

	l.Info("Local latest version", esl.String("version", localVersionLatest.String()), esl.String("path", localVersionPaths[localVersionLatest.String()]))
	l.Info("Remote latest version", esl.String("version", remoteVersionLatest.String()), esl.String("path", remoteVersionPaths[remoteVersionLatest.String()]))

	dlPath, err := z.Download(remoteVersionLatest, remoteVersionPaths[remoteVersionLatest.String()])
	if err != nil {
		l.Warn("Unable to download", esl.Error(err))
		return err
	}
	cellarPath, err := z.Extract(remoteVersionLatest, dlPath)
	if err != nil {
		l.Warn("Unable to extract", esl.Error(err))
		return err
	}
	l.Info("Extracted", esl.String("path", cellarPath))

	return nil
}

func (z binSrcDropboxDstLocalWorkerImpl) GetLocalLatest() (binaryPath string, version es_version.Version, err error) {
	l := z.ctl.Log()
	localVersions, localVersionPaths, err := z.ListLocalVersions()
	if err != nil {
		l.Debug("Unable to list local versions", esl.Error(err))
		return "", es_version.Zero(), err
	}
	localVersionLatest := es_version.Max(localVersions...)
	if len(localVersions) < 1 || localVersionLatest.Equals(es_version.Zero()) {
		if err := z.UpdateIfRequired(); err != nil {
			l.Debug("Unable to update", esl.Error(err))
			return "", es_version.Zero(), err
		}
	}
	return localVersionPaths[localVersionLatest.String()], localVersionLatest, nil
}

func (z binSrcDropboxDstLocalWorkerImpl) ListLocalVersions() (versions []es_version.Version, versionPaths map[string]string, err error) {
	versions = make([]es_version.Version, 0)
	versionPaths = make(map[string]string)
	l := z.ctl.Log().With(esl.String("cellarPath", z.recipe.CellarPath))
	entries, err := os.ReadDir(z.recipe.CellarPath)
	if err != nil {
		if os.IsNotExist(err) {
			l.Debug("Cellar directory not found")
			return versions, versionPaths, nil
		}
		l.Debug("Unable to read cellar directory", esl.Error(err))
		return versions, versionPaths, err
	}
	fullPrefix := z.recipe.Prefix + "-"

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), fullPrefix) {
			verStr := strings.TrimPrefix(entry.Name(), fullPrefix)
			ver, err := es_version.Parse(verStr)
			if err != nil {
				l.Debug("Unable to parse version", esl.Error(err))
				continue
			}
			versions = append(versions, ver)
			versionPaths[ver.String()] = filepath.Join(z.recipe.CellarPath, entry.Name())
		}
	}

	return versions, versionPaths, nil
}

func (z binSrcDropboxDstLocalWorkerImpl) ListRemoteVersions() (versions []es_version.Version, versionPaths map[string]string, err error) {
	l := z.ctl.Log().With(esl.String("sourceUrl", z.recipe.SourceUrl))
	versions = make([]es_version.Version, 0)
	versionPaths = make(map[string]string)

	url, err := mo_url.NewUrl(z.recipe.SourceUrl)
	if err != nil {
		l.Debug("Unable to parse url", esl.Error(err))
		return versions, versionPaths, err
	}
	fullPrefix := z.recipe.Prefix + "-"
	fullFileSuffix := "-" + z.recipe.Suffix + ".zip"

	svs := sv_sharedlink_file.New(z.client)
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

func (z binSrcDropboxDstLocalWorkerImpl) Download(version es_version.Version, versionPath string) (downloadPath string, err error) {
	l := z.ctl.Log().With(esl.String("version", version.String()), esl.String("versionPath", versionPath))
	l.Debug("Download version")

	svs := sv_sharedlink_file.New(z.client)
	url, err := mo_url.NewUrl(z.recipe.SourceUrl)
	if err != nil {
		l.Debug("Unable to parse url", esl.Error(err))
		return "", err
	}
	entry, path, err := svs.Download(
		url,
		dbx_path.NewDropboxPath(versionPath),
		mo_path.NewFileSystemPath(z.recipe.DeployPath),
		sv_sharedlink_file.Password(z.recipe.SourcePassword))
	if err != nil {
		l.Debug("Unable to download version", esl.Error(err))
		return "", err
	}
	l.Info("Downloaded", esl.String("path", path.Path()), esl.String("entry", entry.Path().Path()))

	return path.Path(), nil
}

func (z binSrcDropboxDstLocalWorkerImpl) Extract(version es_version.Version, downloadPath string) (cellarPath string, err error) {
	l := z.ctl.Log().With(esl.String("downloadPath", downloadPath))
	l.Debug("Extract version")

	cellarPath = filepath.Join(z.recipe.CellarPath, z.recipe.Prefix+"-"+version.String())
	if err := os.MkdirAll(cellarPath, 0755); err != nil {
		l.Debug("Unable to create destination directory", esl.Error(err))
		return "", err
	}
	l.Info("Extracting into cellar directory", esl.String("cellarPath", cellarPath))
	err = es_zip.Extract(l, downloadPath, cellarPath)
	l.Debug("Extracted", esl.Error(err), esl.String("cellarPath", cellarPath))

	binCellarPath := filepath.Join(cellarPath, z.BinaryName())

	if err := os.Chmod(binCellarPath, 0755); err != nil {
		l.Warn("Unable to change permission", esl.Error(err))
		return cellarPath, err
	}

	if err := os.Remove(downloadPath); err != nil {
		l.Warn("Unable to remove downloaded file", esl.Error(err))
	}

	return cellarPath, err
}

func (z binSrcDropboxDstLocalWorkerImpl) DeploySymlink() (err error) {
	l := z.ctl.Log()
	l.Info("Deploying symlink")

	versionPath, version, err := z.GetLocalLatest()
	if err != nil {
		l.Warn("Unable to get local latest", esl.Error(err))
		return err
	}

	binName := z.BinaryName()
	binCellarPath := filepath.Join(versionPath, binName)
	binDeployPath := filepath.Join(z.recipe.DeployPath, binName)

	l.Info("Deploying",
		esl.String("version", version.String()),
		esl.String("binCellarPath", binCellarPath),
		esl.String("binDeployPath", binDeployPath))

	_, err = os.Lstat(binDeployPath)
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
	if err := os.MkdirAll(z.recipe.DeployPath, 0755); err != nil {
		l.Warn("Unable to create deploy directory", esl.Error(err))
		return err
	}

	if err := os.Symlink(binCellarPath, binDeployPath); err != nil {
		l.Warn("Unable to create symlink", esl.Error(err))
		return err
	}
	l.Info("Deployed", esl.String("binDeployPath", binDeployPath))
	return nil
}

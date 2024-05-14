package sb_deploy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/watermint/toolbox/domain/dropbox/api/dbx_client"
	"github.com/watermint/toolbox/domain/dropbox/model/mo_file"
	dbx_path "github.com/watermint/toolbox/domain/dropbox/model/mo_path"
	"github.com/watermint/toolbox/domain/dropbox/model/mo_url"
	"github.com/watermint/toolbox/domain/dropbox/service/sv_sharedlink_file"
	"github.com/watermint/toolbox/essentials/log/esl"
	"github.com/watermint/toolbox/essentials/model/mo_path"
	"github.com/watermint/toolbox/essentials/strings/es_version"
	"github.com/watermint/toolbox/infra/control/app_control"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	BinSrcDropboxDstLocalVersionCacheLifecycle = 86400
	BinSrcDropboxDstLocalVersionCacheName      = "sb_deploy-bin_src_dbx_dst_local_version_cache"
)

// BinSrcDropboxDstLocalRecipe Deploy binary from Dropbox to local
// This recipe expect folder structure like this:
// `PREFIX-VERSION/PREFIX-VERSION-SUFFIX.zip`.
// For example, if prefix is `myapp` and suffix is `linux-amd64`, the folder structure is:
// `myapp-1.0.0/myapp-1.0.0-linux-amd64.zip`.
type BinSrcDropboxDstLocalRecipe struct {
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

type BinSrcDropboxDstLocalRemoteVersionCache struct {
	// CacheTime is the time when the cache is created in Unix time
	CacheTime int64 `json:"cache_time,omitempty"`

	// Versions is the list of versions
	Versions []es_version.Version `json:"versions,omitempty"`

	// Versions is the list of versions, version string as key and path as value
	VersionPaths map[string]string `json:"version_paths,omitempty"`
}

func NewBinSrcDropboxDstLocal(recipe BinSrcDropboxDstLocalRecipe, ctl app_control.Control, client dbx_client.Client) BinDeploy {
	return &binSrcDropboxDstLocalWorkerImpl{
		recipe: recipe,
		ctl:    ctl,
		client: client,
	}
}

type binSrcDropboxDstLocalWorkerImpl struct {
	recipe BinSrcDropboxDstLocalRecipe
	ctl    app_control.Control
	client dbx_client.Client
}

func (z binSrcDropboxDstLocalWorkerImpl) IsUpdateRequired() (required bool, err error) {
	localVersions, _, err := z.ListLocalVersions()
	if err != nil {
		return false, err
	}
	remoteVersions, _, err := z.ListRemoteVersions()
	if err != nil {
		return false, err
	}

	remoteVersionLatest := es_version.Max(remoteVersions...)
	localVersionLatest := es_version.Max(localVersions...)

	return !localVersionLatest.Equals(remoteVersionLatest), nil
}

func (z binSrcDropboxDstLocalWorkerImpl) LocalLatestBinaryPath() string {
	return utilLocalLatestBinaryPath(z.ctl, z.recipe.CellarPath, z.recipe.Prefix, z.recipe.BinaryName)
}

func (z binSrcDropboxDstLocalWorkerImpl) BinaryName() string {
	return utilBinaryName(z.recipe.BinaryName)
}

func (z binSrcDropboxDstLocalWorkerImpl) update(force bool) (err error) {
	l := z.ctl.Log()

	updateRequired, err := z.IsUpdateRequired()
	if err != nil {
		l.Warn("Unable to check update required", esl.Error(err))
		return err
	}
	if !force && !updateRequired {
		l.Info("No update required")
		return nil
	}

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

func (z binSrcDropboxDstLocalWorkerImpl) UpdateForce() (err error) {
	return z.update(true)
}

func (z binSrcDropboxDstLocalWorkerImpl) UpdateIfRequired() (err error) {
	return z.update(false)
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

func (z binSrcDropboxDstLocalWorkerImpl) remoteVersionCacheName() string {
	seeds := make([]string, 0)
	seeds = append(seeds, z.recipe.SourceUrl)
	seeds = append(seeds, z.recipe.Prefix)
	seeds = append(seeds, z.recipe.Suffix)
	seeds = append(seeds, z.recipe.BinaryName)
	seed := sha256.Sum256([]byte(strings.Join(seeds, "-")))
	return BinSrcDropboxDstLocalVersionCacheName + hex.EncodeToString(seed[:])[0:16] + ".json"
}

func (z binSrcDropboxDstLocalWorkerImpl) loadRemoteVersionsCache() (versions []es_version.Version, versionPaths map[string]string, found bool) {
	l := z.ctl.Log()
	cachePath := filepath.Join(z.ctl.Workspace().Cache(), z.remoteVersionCacheName())
	cacheData, err := os.ReadFile(cachePath)
	if err != nil {
		l.Debug("Unable to read cache", esl.Error(err))
		return versions, versionPaths, false
	}
	l.Debug("Cache found")
	cache := &BinSrcDropboxDstLocalRemoteVersionCache{}
	if err = json.Unmarshal(cacheData, cache); err != nil {
		l.Debug("Unable to unmarshal cache", esl.Error(err))
		return versions, versionPaths, false
	}
	if cache.CacheTime+BinSrcDropboxDstLocalVersionCacheLifecycle < time.Now().Unix() {
		return versions, versionPaths, false
	}
	return cache.Versions, cache.VersionPaths, true
}

func (z binSrcDropboxDstLocalWorkerImpl) saveRemoteVersionCache(versions []es_version.Version, versionPaths map[string]string) (err error) {
	l := z.ctl.Log()
	if err := os.MkdirAll(z.ctl.Workspace().Cache(), 0755); err != nil {
		l.Debug("Unable to create cache directory", esl.Error(err))
		return err
	}

	cachePath := filepath.Join(z.ctl.Workspace().Cache(), z.remoteVersionCacheName())
	cache := &BinSrcDropboxDstLocalRemoteVersionCache{
		CacheTime:    time.Now().Unix(),
		Versions:     versions,
		VersionPaths: versionPaths,
	}
	cacheData, err := json.Marshal(cache)
	if err != nil {
		l.Debug("Unable to marshal cache", esl.Error(err))
		return err
	}
	if err := os.WriteFile(cachePath, cacheData, 0644); err != nil {
		l.Debug("Unable to write cache", esl.Error(err))
		return err
	}
	return nil
}

func (z binSrcDropboxDstLocalWorkerImpl) ListLocalVersions() (versions []es_version.Version, versionPaths map[string]string, err error) {
	return utilLocalListLocalVersions(z.ctl, z.recipe.CellarPath, z.recipe.Prefix)
}

func (z binSrcDropboxDstLocalWorkerImpl) ListRemoteVersions() (versions []es_version.Version, versionPaths map[string]string, err error) {
	l := z.ctl.Log().With(esl.String("sourceUrl", z.recipe.SourceUrl))
	versions = make([]es_version.Version, 0)
	versionPaths = make(map[string]string)

	if versions, versionPaths, found := z.loadRemoteVersionsCache(); found {
		l.Debug("Remote version cache found")
		return versions, versionPaths, nil
	}

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

	if err := z.saveRemoteVersionCache(versions, versionPaths); err != nil {
		l.Debug("Unable to save remote version cache", esl.Error(err))
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
	return utilLocalExtract(z.ctl, z.recipe.CellarPath, z.recipe.Prefix, z.recipe.BinaryName, version, downloadPath)
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

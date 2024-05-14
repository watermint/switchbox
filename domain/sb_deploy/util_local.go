package sb_deploy

import (
	"github.com/watermint/toolbox/essentials/io/es_zip"
	"github.com/watermint/toolbox/essentials/log/esl"
	"github.com/watermint/toolbox/essentials/strings/es_version"
	"github.com/watermint/toolbox/infra/control/app_control"
	"github.com/watermint/toolbox/infra/control/app_definitions"
	"os"
	"path/filepath"
	"strings"
)

func utilBinaryName(binName string) string {
	if app_definitions.IsWindows() {
		return binName + ".exe"
	} else {
		return binName
	}
}

func utilLocalExtract(c app_control.Control, cellarPath, prefix, binName string, version es_version.Version, downloadPath string) (versionCellarPath string, err error) {
	l := c.Log().With(esl.String("downloadPath", downloadPath))
	l.Debug("Extract version")

	versionCellarPath = filepath.Join(cellarPath, prefix+"-"+version.String())
	if err := os.MkdirAll(versionCellarPath, 0755); err != nil {
		l.Debug("Unable to create destination directory", esl.Error(err))
		return "", err
	}
	l.Info("Extracting into cellar directory", esl.String("cellarPath", versionCellarPath))
	err = es_zip.Extract(l, downloadPath, versionCellarPath)
	l.Debug("Extracted", esl.Error(err), esl.String("cellarPath", versionCellarPath))

	binCellarPath := filepath.Join(versionCellarPath, utilBinaryName(binName))

	if err := os.Chmod(binCellarPath, 0755); err != nil {
		l.Warn("Unable to change permission", esl.Error(err))
		return versionCellarPath, err
	}

	if err := os.Remove(downloadPath); err != nil {
		l.Warn("Unable to remove downloaded file", esl.Error(err))
	}

	return versionCellarPath, err
}

func utilLocalLatestBinaryPath(c app_control.Control, cellarPath, prefix, binName string) string {
	l := c.Log()
	localVersions, localVersionPaths, err := utilLocalListLocalVersions(c, cellarPath, prefix)
	if err != nil {
		l.Debug("Unable to list local versions", esl.Error(err))
		return ""
	}
	localVersionLatest := es_version.Max(localVersions...)
	if len(localVersions) < 1 || localVersionLatest.Equals(es_version.Zero()) {
		return ""
	}
	return filepath.Join(localVersionPaths[localVersionLatest.String()], utilBinaryName(binName))
}

func utilLocalListLocalVersions(c app_control.Control, cellarPath, prefix string) (versions []es_version.Version, versionPaths map[string]string, err error) {
	versions = make([]es_version.Version, 0)
	versionPaths = make(map[string]string)
	l := c.Log().With(esl.String("cellarPath", cellarPath))
	entries, err := os.ReadDir(cellarPath)
	if err != nil {
		if os.IsNotExist(err) {
			l.Debug("Cellar directory not found")
			return versions, versionPaths, nil
		}
		l.Debug("Unable to read cellar directory", esl.Error(err))
		return versions, versionPaths, err
	}
	fullPrefix := prefix + "-"

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), fullPrefix) {
			verStr := strings.TrimPrefix(entry.Name(), fullPrefix)
			ver, err := es_version.Parse(verStr)
			if err != nil {
				l.Debug("Unable to parse version", esl.Error(err))
				continue
			}
			versions = append(versions, ver)
			versionPaths[ver.String()] = filepath.Join(cellarPath, entry.Name())
		}
	}

	return versions, versionPaths, nil
}

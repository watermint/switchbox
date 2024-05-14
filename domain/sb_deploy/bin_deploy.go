package sb_deploy

import "github.com/watermint/toolbox/essentials/strings/es_version"

type BinDeploy interface {
	// UpdateIfRequired Update if required
	UpdateIfRequired() (err error)

	// UpdateForce Update forcefully
	UpdateForce() (err error)

	// IsUpdateRequired Check if update is required
	IsUpdateRequired() (required bool, err error)

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

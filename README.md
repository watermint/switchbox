# watermint switchbox

[![Build](https://github.com/watermint/toolbox/actions/workflows/build.yml/badge.svg)](https://github.com/watermint/toolbox/actions/workflows/build.yml)
[![Test](https://github.com/watermint/toolbox/actions/workflows/test.yml/badge.svg)](https://github.com/watermint/toolbox/actions/workflows/test.yml)
[![CodeQL](https://github.com/watermint/toolbox/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/watermint/toolbox/actions/workflows/codeql-analysis.yml)
[![Codecov](https://codecov.io/gh/watermint/toolbox/branch/main/graph/badge.svg?token=CrE8reSVvE)](https://codecov.io/gh/watermint/toolbox)

![watermint toolbox](resources/images/watermint-toolbox-256x256.png)

The multi-purpose utility command-line tool for web services including Dropbox, Dropbox for teams, Google, GitHub, etc.

# Licensing & Disclaimers

watermint switchbox is licensed under the Apache License, Version 2.0.
Please see LICENSE.md or LICENSE.txt for more detail.

Please carefully note:
> Unless required by applicable law or agreed to in writing, Licensor provides the Work (and each Contributor provides its Contributions) on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied, including, without limitation, any warranties or conditions of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A PARTICULAR PURPOSE.

# Built executable

Pre-compiled binaries can be found in [Latest Release](https://github.com/watermint/toolbox/releases/latest). If you are building directly from the source, please refer [BUILD.md](BUILD.md).

## Installing using Homebrew on macOS/Linux

First, you need to install Homebrew. Please refer the instruction on [the official site](https://brew.sh/). Then run following commands to install watermint toolbox.
```
brew tap watermint/toolbox
brew install toolbox
```

# Product lifecycle

## Maintenance policy

This product itself is experimental and is not subject to the maintained to keep quality of service. The project will try to fix critical bugs and security issues with the best effort. But that is also not guaranteed.

The product will not release any patch release of a certain major releases. The product will apply fixes as next release when those fixes accepted to do.

## Specification changes

The deliverables of this project are stand-alone executable programs. The specification changes will not be applied unless you explicitly upgrade your version of the program.

The following policy will be used to make changes in new version releases.

Command paths, arguments, return values, etc. will be upgraded to be as compatible as possible, but may be discontinued or changed.Addition of arguments, etc.
The general policy is as follows.

* Changes that do not break existing behavior, such as the addition of arguments or changes to messages, will be implemented without notice.
* Commands that are considered infrequently used will be discontinued or moved without notice.
* Changes to other commands will be announced 30-180 days or more in advance.

Changes in specifications will be announced at [Announcements](https://github.com/watermint/toolbox/discussions/categories/announcements). Please refer to [Specification Change](https://toolbox.watermint.org/guides/spec-change.html) for a list of planned specification changes.

## Availability period for each release

In general, new security issues are discovered every day. In order not to leave these security and critical issues unaddressed by continuing to use older watermint toolbox releases, a maximum availability period has been set for release 130 and above. Please see [#815](https://github.com/watermint/toolbox/discussions/815) for more details.

# Announcements

# Security and privacy

## Information Not Collected 

The watermint toolbox does not collect any information to third-party servers.

The watermint toolbox is for interacting with the services like Dropbox with your account. There is no third-party account involved. The Commands stores API token, logs, files, or reports on your PC storage.

## Sensitive data

Most sensitive data, such as API token, are saved on your PC storage in obfuscated & made restricted access. However, it's your responsibility to keep those data secret. 
Please do not share files, especially the `secrets` folder under toolbox work path (`C:\Users\<your user name>\.toolbox`, or `$HOME/.toolbox` by default).

## Internet access other than the service that is the subject of the command used

The watermint toolbox has the ability to deactivate certain releases that have critical bugs or security issues. It does this by retrieving data from a repository hosted on GitHub about once every 30 days to check the status of a release.
This access does not collect your private data (such as your Dropbox, Google, local files, token, etc). It just checks the release status, but as a side effect your IP address is sent to GitHub when downloading data. I know IP address is also a PII. But this is the same as visiting a general website and is not a special operation.
The watermint toolbox project administrator cannot even see how many times these files have been downloaded.

# Usage

`tbx` have various features. Run without an option for a list of supported commands and options.
You can see available commands and options by running executable without arguments like below.

```
% ./sbx

watermint switchbox xx.x.xxx
============================

© 2024-2024 Takayuki Okazaki
Licensed under open source licenses. Use the `license` command for more detail.

Tools for Dropbox and Dropbox for teams

Usage:
======

./sbx  command

Available commands:
===================

| Command  | Description              | Notes |
|----------|--------------------------|-------|
| config   | CLI configuration        |       |
| deploy   | Deploy commands          |       |
| dispatch | Dispatch commands        |       |
| license  | Show license information |       |
| version  | Show version             |       |

```

# Commands

## Dropbox (Individual account)

| Command                                         | Description                                                                   |
|-------------------------------------------------|-------------------------------------------------------------------------------|
| [deploy link](docs/commands/deploy-link.md)     | Deploy binary from Dropbox shared link and create symbolic link to the binary |
| [deploy update](docs/commands/deploy-update.md) | Update binary from Dropbox shared link                                        |
| [dispatch run](docs/commands/dispatch-run.md)   | Run the latest version of the binary                                          |

## GitHub

| Command                                                     | Description              |
|-------------------------------------------------------------|--------------------------|
| [dev release checkin](docs/commands/dev-release-checkin.md) | Check in the new release |

## Utilities

| Command                                                           | Description                       |
|-------------------------------------------------------------------|-----------------------------------|
| [config auth delete](docs/commands/config-auth-delete.md)         | Delete existing auth credential   |
| [config auth list](docs/commands/config-auth-list.md)             | List all auth credentials         |
| [config feature disable](docs/commands/config-feature-disable.md) | Disable a feature.                |
| [config feature enable](docs/commands/config-feature-enable.md)   | Enable a feature.                 |
| [config feature list](docs/commands/config-feature-list.md)       | List available optional features. |
| [config license list](docs/commands/config-license-list.md)       | List available license keys       |
| [license](docs/commands/license.md)                               | Show license information          |
| [version](docs/commands/version.md)                               | Show version                      |


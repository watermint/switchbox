package gb_definitions

import (
	"github.com/watermint/toolbox/infra/control/app_definitions"
	"time"
)

const (
	PkgBase = "github.com/watermint/switchbox"

	LifecycleExpirationWarning  = 365 * 24 * time.Hour // 1 year
	LifecycleExpirationCritical = 365 * 24 * time.Hour // 1 year
	LifecycleExpirationMode     = app_definitions.LifecycleExpirationShutdown
	LifecycleUpgradeUrl         = "https://github.com/watermint/switchbox"
)

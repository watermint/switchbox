package sb_recipe

import "github.com/watermint/toolbox/infra/control/app_control"

// Preset is a recipe preset interface.
type Preset interface {
	Preset()
}

// Recipe is called before Exec() and Test().
// This interface definition required to generate recipe catalogue.
// This interface need to be same as rc_recipe.Recipe.
type Recipe interface {
	Preset
	Exec(c app_control.Control) error
	Test(c app_control.Control) error
}

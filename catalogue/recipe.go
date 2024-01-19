package catalogue

// Code generated by dev catalogue command DO NOT EDIT

import (
	infra_recipe_rc_recipe "github.com/watermint/switchbox/infra/sb_recipe"
	recipedeploy "github.com/watermint/switchbox/recipe/deploy"
	recipedispatch "github.com/watermint/switchbox/recipe/dispatch"
)

func AutoDetectedRecipes() []infra_recipe_rc_recipe.Recipe {
	return []infra_recipe_rc_recipe.Recipe{
		&recipedeploy.Update{},
		&recipedispatch.Run{},
	}
}

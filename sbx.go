package main

import (
	"fmt"
	switchboxcatalogue "github.com/watermint/switchbox/catalogue"
	sb_definitions "github.com/watermint/switchbox/infra/sb_definitions"
	"github.com/watermint/switchbox/resources"
	toolboxcatalogue "github.com/watermint/toolbox/catalogue"
	"github.com/watermint/toolbox/essentials/log/esl"
	"github.com/watermint/toolbox/essentials/log/wrapper/lgw_golog"
	"github.com/watermint/toolbox/infra/control/app_bootstrap"
	"github.com/watermint/toolbox/infra/control/app_build"
	"github.com/watermint/toolbox/infra/control/app_catalogue"
	"github.com/watermint/toolbox/infra/control/app_definitions"
	"github.com/watermint/toolbox/infra/control/app_exit"
	"github.com/watermint/toolbox/infra/control/app_feature"
	"github.com/watermint/toolbox/infra/control/app_resource"
	"github.com/watermint/toolbox/infra/doc/dc_supplemental"
	"github.com/watermint/toolbox/infra/recipe/rc_catalogue"
	"github.com/watermint/toolbox/infra/recipe/rc_catalogue_impl"
	"github.com/watermint/toolbox/infra/recipe/rc_recipe"
	"github.com/watermint/toolbox/infra/recipe/rc_spec"
	toolboxresources "github.com/watermint/toolbox/resources"
	"log"
	"os"
	"strings"
)

var (
	ImportToolboxRecipesPrefixes = []string{
		"config",
		"dev",
		"license",
		"version",
	}
)

func loadCatalogue() rc_catalogue.Catalogue {
	recipes := make([]rc_recipe.Recipe, 0)
	toolboxRecipes := make([]rc_recipe.Recipe, 0)
	toolboxRecipes = append(toolboxRecipes, toolboxcatalogue.AutoDetectedRecipesClassic()...)
	toolboxRecipes = append(toolboxRecipes, toolboxcatalogue.AutoDetectedRecipesCitron()...)
	for _, r := range toolboxRecipes {
		rs := rc_spec.New(r)
		for _, p := range ImportToolboxRecipesPrefixes {
			if strings.HasPrefix(rs.CliPath(), p) {
				recipes = append(recipes, r)
				break
			}
		}
	}
	for _, r := range switchboxcatalogue.AutoDetectedRecipes() {
		recipes = append(recipes, r)
	}
	ingredients := make([]rc_recipe.Recipe, 0)
	//ingredients = append(ingredients, switchboxcatalogue.AutoDetectedIngredients()...)
	ingredients = append(ingredients, toolboxcatalogue.AutoDetectedIngredients()...)
	messages := make([]interface{}, 0)
	messages = append(messages, switchboxcatalogue.AutoDetectedMessageObjects()...)
	messages = append(messages, toolboxcatalogue.AutoDetectedMessageObjects()...)
	features := make([]app_feature.OptIn, 0)
	features = append(features, switchboxcatalogue.AutoDetectedFeatures()...)
	features = append(features, toolboxcatalogue.AutoDetectedFeatures()...)

	return rc_catalogue_impl.NewCatalogue(recipes, ingredients, messages, features)
}

func run(args []string, testMode bool) {
	defer func() {
		if r := recover(); r != nil {
			if r == app_exit.Success {
				return
			} else {
				panic(r)
			}
		}
	}()

	// settings
	app_definitions.PackagesBaseKey = []string{
		app_definitions.CorePkg,
		sb_definitions.PkgBase,
	}
	app_definitions.PackagesBaseRecipe = []string{
		app_definitions.CorePkg + "/recipe",
		app_definitions.CorePkg + "/citron",
		sb_definitions.PkgBase + "/recipe",
	}
	dc_supplemental.SkipDropboxBusinessCommandDoc = true

	// identify resource bundle
	bundle := resources.NewMergedBundle()
	toolboxresources.CurrentBundle = bundle
	app_build.CurrentRelease = toolboxresources.Release()
	build := toolboxresources.BuildFromResource(bundle.Build())
	resolvedVersion := app_build.SelectVersion(build.Version)

	// Override build info
	app_definitions.Name = "watermint switchbox"
	app_definitions.ExecutableName = "sbx"
	app_definitions.Version = resolvedVersion
	app_definitions.BuildInfo = build
	app_definitions.BuildId = resolvedVersion.String()
	app_definitions.Release = toolboxresources.Release()
	app_definitions.ApplicationRepositoryName = "switchbox"
	app_definitions.ApplicationRepositoryOwner = "watermint"
	app_definitions.Copyright = fmt.Sprintf("© 2024-%4d Takayuki Okazaki", build.Year)
	app_definitions.LandingPage = "https://github.com/watermint/switchbox"
	app_definitions.LifecycleExpirationWarning = sb_definitions.LifecycleExpirationWarning
	app_definitions.LifecycleExpirationCritical = sb_definitions.LifecycleExpirationCritical
	app_definitions.LifecycleUpgradeUrl = sb_definitions.LifecycleUpgradeUrl
	app_definitions.RecipePackageNames = []string{
		"recipe",
	}

	app_exit.SetTestMode(testMode)

	app_resource.SetBundle(bundle)
	app_catalogue.SetCurrent(loadCatalogue())
	log.SetOutput(lgw_golog.NewLogWrapper(esl.Default()))

	b := app_bootstrap.NewBootstrap()
	b.Run(b.Parse(args...))
}

func main() {
	run(os.Args[1:], false)
}

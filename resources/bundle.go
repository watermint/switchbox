package resources

import (
	"github.com/watermint/toolbox/essentials/go/es_lang"
	"github.com/watermint/toolbox/essentials/go/es_resource"
)
import toolboxresources "github.com/watermint/toolbox/resources"

// NewCurrentBundle returns the current project bundle.
func NewCurrentBundle() es_resource.Bundle {
	return es_resource.New(
		es_resource.NewResource("templates", resTemplates),
		es_resource.NewNonTraversableResource("messages", resMessages),
		es_resource.NewResource("web", resWeb),
		es_resource.NewNonTraversableResource("keys", resKeys),
		es_resource.NewResource("images", resImages),
		es_resource.NewNonTraversableResource("data", resData),
		es_resource.NewNonTraversableResource("build", resBuildInfo),
		es_resource.NewNonTraversableResource("release", resRelease),
	)
}

// NewMergedBundle returns the merged bundle of the current project and toolbox.
func NewMergedBundle() es_resource.Bundle {
	langCodes := make([]string, 0)
	for _, l := range es_lang.Supported {
		langCodes = append(langCodes, l.CodeString())
	}
	return es_resource.NewChainBundle(langCodes, NewCurrentBundle(), toolboxresources.NewBundle())
}

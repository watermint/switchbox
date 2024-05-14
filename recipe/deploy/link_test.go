package deploy

import (
	"github.com/watermint/toolbox/quality/recipe/qtr_endtoend"
	"testing"
)

func TestLink_Exec(t *testing.T) {
	qtr_endtoend.TestRecipe(t, &Link{})
}

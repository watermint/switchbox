package deploy

import (
	"github.com/watermint/toolbox/quality/recipe/qtr_endtoend"
	"testing"
)

func TestBin_Exec(t *testing.T) {
	qtr_endtoend.TestRecipe(t, &Bin{})
}

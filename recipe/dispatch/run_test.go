package dispatch

import (
	"github.com/watermint/toolbox/quality/recipe/qtr_endtoend"
	"testing"
)

func TestRun_Exec(t *testing.T) {
	qtr_endtoend.TestRecipe(t, &Run{})
}

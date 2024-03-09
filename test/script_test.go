package test

import (
	"os"
	"testing"

	"github.com/SiasMey/notebox/pkg/nbx"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"nbx": nbx.Main,
	}))
}

func Test(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

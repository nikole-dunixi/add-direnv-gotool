package cli_test

import (
	"slices"
	"testing"

	"github.com/nikole-dunixi/add-direnv-gotool/internal/cli"
	"github.com/stretchr/testify/assert"
)

func TestDetermineCommandName(t *testing.T) {
	type testcase struct {
		path          string
		expectedValue string
		expectedOK    bool
	}
	testcases := []testcase{
		// valid cases
		{
			path:          "github.com/onsi/ginkgo/v2/ginkgo@latest",
			expectedValue: "ginkgo",
			expectedOK:    true,
		},
		{
			path:          "gitlab.com/gitlab-org/cli/cmd/glab@main",
			expectedValue: "glab",
			expectedOK:    true,
		},
		{
			path:          "github.com/mitranim/gow@latest",
			expectedValue: "gow",
			expectedOK:    true,
		},
		{
			path:          "golang.org/x/tools/cmd/goimports@latest",
			expectedValue: "goimports",
			expectedOK:    true,
		},
		{
			path:          "github.com/golangci/golangci-lint/v2/cmd/golangci-lint",
			expectedValue: "golangci-lint",
			expectedOK:    true,
		},
		{
			path:          "github.com/magefile/mage@latest",
			expectedValue: "mage",
			expectedOK:    true,
		},
		{
			path:          "github.com/ends/with/slash/",
			expectedValue: "slash",
			expectedOK:    true,
		},
		{
			path:          "github.com/go-task/task/v3/cmd/task@latest",
			expectedValue: "task",
			expectedOK:    true,
		},
		{
			path:          "github.com/a-h/templ/cmd/templ@latest",
			expectedValue: "templ",
			expectedOK:    true,
		},
		{
			path:          "github.com/templui/templui/cmd/templui@latest",
			expectedValue: "templui",
			expectedOK:    true,
		},
		// invalid cases
		{
			path:          "@latest",
			expectedValue: "",
			expectedOK:    false,
		},
	}
	for tc := range slices.Values(testcases) {
		t.Run(tc.path+"=>"+tc.expectedValue, func(t *testing.T) {
			commandName, ok := cli.DetermineCommandName(tc.path)
			assert.Equal(t, tc.expectedOK, ok)
			assert.Equal(t, tc.expectedValue, commandName)
		})
	}
}

package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tufin/oasdiff/checker"
)

func TestGetGitHubActionResponse(t *testing.T) {

	response, err := GetGitHubActionResponse(checker.Changes{checker.ApiChange{
		Level: checker.ERR,
	}, checker.ApiChange{
		Level: checker.INFO,
	}, checker.ApiChange{
		Level: checker.WARN,
	}, checker.ApiChange{
		Level: checker.ERR,
	}})
	require.NoError(t, err)
	require.Equal(t, "4 changes: 2 error, 1 warning, 1 info", response.Summary)
}

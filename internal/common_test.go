package internal_test

import (
	"net/http"
	"testing"

	"github.com/oasdiff/oasdiff-service/internal"
	"github.com/stretchr/testify/require"
)

func TestCreateConfig_PathFilter(t *testing.T) {

	const expected = "test"

	r := createMockRequest(t)
	q := r.URL.Query()
	q.Add("path-filter", expected)
	r.URL.RawQuery = q.Encode()

	config := internal.CreateConfig(r)

	require.Equal(t, expected, config.MatchPath)
}

func createMockRequest(t *testing.T) *http.Request {

	res, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	return res
}

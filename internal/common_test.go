package internal_test

import (
	"net/http"
	"testing"

	"github.com/oasdiff/oasdiff-service/internal"
	"github.com/stretchr/testify/require"
)

func TestCreateConfig_ExcludeExamples(t *testing.T) {

	r := createMockRequest(t)
	q := r.URL.Query()
	q.Add("exclude-examples", "true")
	r.URL.RawQuery = q.Encode()

	config := internal.CreateConfig(r)

	require.Equal(t, true, config.ExcludeExamples)
}

func TestCreateConfig_PathFilter(t *testing.T) {

	const expected = "test"

	r := createMockRequest(t)
	q := r.URL.Query()
	q.Add("path-filter", expected)
	r.URL.RawQuery = q.Encode()

	config := internal.CreateConfig(r)

	require.Equal(t, expected, config.PathFilter)
}

func TestCreateConfig_DeprecationDays(t *testing.T) {

	r := createMockRequest(t)
	q := r.URL.Query()
	q.Add("deprecation-days", "3")
	r.URL.RawQuery = q.Encode()

	config := internal.CreateConfig(r)

	require.Equal(t, 3, config.DeprecationDays)
}

func createMockRequest(t *testing.T) *http.Request {

	res, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	return res
}

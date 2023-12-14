package badge_test

import (
	"testing"

	"github.com/oasdiff/oasdiff-service/internal/badge"
	"github.com/stretchr/testify/require"
)

func TestGenerator_GenerateFlat(t *testing.T) {

	bg, err := badge.NewGenerator("verdana.ttf", 11)
	require.NoError(t, err)
	require.NotEmpty(t, bg.Flat("changelog", "v1.23.4", badge.COLOR_BLUE))
}

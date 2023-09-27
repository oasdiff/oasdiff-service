package internal_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/oasdiff/oasdiff-service/internal"
	"github.com/oasdiff/telemetry/client"
	"github.com/stretchr/testify/require"
	"github.com/tufin/oasdiff/checker"
	"gopkg.in/yaml.v3"
)

func TestBreakingChanges(t *testing.T) {

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	basePart, err := writer.CreateFormFile("base", "openapi-test1.yaml")
	require.NoError(t, err)
	base, err := os.Open("../data/openapi-test1.yaml")
	require.NoError(t, err)
	_, err = io.Copy(basePart, base)
	require.NoError(t, err)

	revisionPart, err := writer.CreateFormFile("revision", "openapi-test3.yaml")
	require.NoError(t, err)
	revision, err := os.Open("../data/openapi-test3.yaml")
	require.NoError(t, err)
	_, err = io.Copy(revisionPart, revision)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	r, err := http.NewRequest(http.MethodPost, "/breaking-changes", body)
	require.NoError(t, err)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		called = true
	}))
	c := client.NewDefaultCollector()
	c.EventsUrl = server.URL

	internal.NewHandler(c).BreakingChangesFromFile(w, r)

	require.Equal(t, http.StatusCreated, w.Result().StatusCode)
	var report map[string][]checker.ApiChange
	require.NoError(t, yaml.NewDecoder(w.Result().Body).Decode(&report))
	require.True(t, len(report["breaking-changes"]) > 0)
	require.True(t, called)
}

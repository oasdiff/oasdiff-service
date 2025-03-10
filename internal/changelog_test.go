package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/oasdiff/oasdiff-service/internal"
	"github.com/oasdiff/telemetry/client"
	"github.com/oasdiff/telemetry/model"
	"github.com/stretchr/testify/require"
	"github.com/oasdiff/oasdiff/formatters"
	"gopkg.in/yaml.v3"
)

func TestChangelog(t *testing.T) {

	const headerUserAgent = "go-test"

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

	r, err := http.NewRequest(http.MethodPost, "/changelog", body)
	require.NoError(t, err)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("User-Agent", headerUserAgent)
	w := httptest.NewRecorder()

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		var events map[string][]*model.Telemetry
		require.NoError(t, json.NewDecoder(r.Body).Decode(&events))
		telemetry := events[model.KeyEvents][0]
		require.Equal(t, headerUserAgent, telemetry.Platform)
		called = true
	}))
	c := client.NewDefaultCollector()
	c.EventsUrl = server.URL

	internal.NewHandler(c).ChangelogFromFile(w, r)

	require.Equal(t, http.StatusCreated, w.Result().StatusCode)
	var report map[string][]formatters.Change
	require.NoError(t, yaml.NewDecoder(w.Result().Body).Decode(&report))
	require.True(t, len(report["changes"]) > 0)
	require.True(t, called)
}

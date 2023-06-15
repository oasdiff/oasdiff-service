package internal

import (
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"
	"github.com/tufin/oasdiff/report"
	"gopkg.in/yaml.v3"
)

func DiffFromUri(w http.ResponseWriter, r *http.Request) {

	base := getQueryString(r, "base", "")
	if base == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	revision := getQueryString(r, "revision", "")
	if revision == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	baseSpec, revisionSpec, code := createSpecFromUri(base, revision)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	createDiffReport(w, r, baseSpec, revisionSpec)
}

func DiffFromFile(w http.ResponseWriter, r *http.Request) {

	dir, base, revision, code := CreateFiles(r)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}
	defer CloseFile(base)
	defer CloseFile(revision)
	defer os.RemoveAll(dir)

	baseSpec, revisionSpec, code := createSpecFromFile(base, revision)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	createDiffReport(w, r, baseSpec, revisionSpec)
}

func createSpecFromUri(base, revision string) (*openapi3.T, *openapi3.T, int) {
	panic("unimplemented")
}

func createSpecFromFile(base *os.File, revision *os.File) (*openapi3.T, *openapi3.T, int) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false

	s1, err := loader.LoadFromFile(base.Name())
	if err != nil {
		log.Infof("failed to load base spec from %q with %v", base.Name(), err)
		return nil, nil, http.StatusBadRequest
	}

	s2, err := load.From(loader, revision.Name())
	if err != nil {
		log.Infof("failed to load revision spec from %q with %v", revision.Name(), err)
		return nil, nil, http.StatusBadRequest
	}

	return s1, s2, http.StatusOK
}

func createDiffReport(w http.ResponseWriter, r *http.Request, s1 *openapi3.T, s2 *openapi3.T) {

	config := CreateConfig(r)

	diffReport, err := diff.Get(config, s1, s2)
	if err != nil {
		log.Infof("failed to calculate diff between a pair of OpenAPI objects '%s' with %v", s1.Info.Title, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Header.Get(HeaderAccept) == HeaderTextHtml {
		html, err := report.GetHTMLReportAsString(diffReport)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Errorf("failed to generate HTML diff report with '%v'", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(html))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set(HeaderContentType, HeaderAppYaml)
	err = yaml.NewEncoder(w).Encode(diffReport)
	if err != nil {
		log.Errorf("failed to encode 'diff' report '%s' with '%v'", s1.Info.Title, err)
	}
}

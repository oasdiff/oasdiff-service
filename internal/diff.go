package internal

import (
	"net/http"
	"net/url"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"
	"github.com/tufin/oasdiff/report"
	"gopkg.in/yaml.v3"
)

func DiffFromUri(w http.ResponseWriter, r *http.Request) {

	base := GetQueryString(r, "base", "")
	if base == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	revision := GetQueryString(r, "revision", "")
	if revision == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	baseSpec, revisionSpec, code := createSpecFromUri(base, revision)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	diffReport, code := createDiffReport(r, baseSpec, revisionSpec)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	// html response
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

	// json response
	w.WriteHeader(http.StatusCreated)
	w.Header().Set(HeaderContentType, HeaderAppYaml)
	err := yaml.NewEncoder(w).Encode(diffReport)
	if err != nil {
		log.Errorf("failed to encode 'diff' report '%s' with '%v'", baseSpec.Info.Title, err)
	}
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

	diffReport, code := createDiffReport(r, baseSpec, revisionSpec)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	// html response
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

	// json response
	w.WriteHeader(http.StatusCreated)
	w.Header().Set(HeaderContentType, HeaderAppYaml)
	err := yaml.NewEncoder(w).Encode(diffReport)
	if err != nil {
		log.Errorf("failed to encode 'diff' report '%s' with '%v'", baseSpec.Info.Title, err)
	}
}

func createSpecFromUri(base string, revision string) (*openapi3.T, *openapi3.T, int) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	u, err := url.Parse(base)
	if err != nil {
		log.Infof("failed to url parse base spec '%s' with '%v'", base, err)
		return nil, nil, http.StatusBadRequest
	}
	s1, err := loader.LoadFromURI(u)
	if err != nil {
		log.Infof("failed to load base spec from '%s' with '%v'", base, err)
		return nil, nil, http.StatusBadRequest
	}

	u, err = url.Parse(revision)
	if err != nil {
		log.Infof("failed to url parse revision spec '%s' with '%v'", revision, err)
		return nil, nil, http.StatusBadRequest
	}
	s2, err := loader.LoadFromURI(u)
	if err != nil {
		log.Infof("failed to load revision spec from '%s' with '%v'", revision, err)
		return nil, nil, http.StatusBadRequest
	}

	return s1, s2, http.StatusOK
}

func createSpecFromFile(base *os.File, revision *os.File) (*openapi3.T, *openapi3.T, int) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false

	s1, err := loader.LoadFromFile(base.Name())
	if err != nil {
		log.Infof("failed to load base spec from '%s' with '%v'", base.Name(), err)
		return nil, nil, http.StatusBadRequest
	}

	s2, err := load.From(loader, revision.Name())
	if err != nil {
		log.Infof("failed to load revision spec from '%s' with '%v'", revision.Name(), err)
		return nil, nil, http.StatusBadRequest
	}

	return s1, s2, http.StatusOK
}

func createDiffReport(r *http.Request, s1 *openapi3.T, s2 *openapi3.T) (*diff.Diff, int) {

	config := CreateConfig(r)

	diffReport, err := diff.Get(config, s1, s2)
	if err != nil {
		log.Infof("failed to calculate diff between a pair of OpenAPI objects '%s' with %v", s1.Info.Title, err)
		return nil, http.StatusBadRequest
	}

	return diffReport, http.StatusOK
}

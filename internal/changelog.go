package internal

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"github.com/tufin/oasdiff/checker"
	"github.com/tufin/oasdiff/checker/localizations"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"
	"gopkg.in/yaml.v3"
)

func ChangelogFromUri(w http.ResponseWriter, r *http.Request) {

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

	changes, code := calcChangelog(r, base, revision)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	res := map[string]checker.Changes{
		"changelog": changes}
	w.WriteHeader(http.StatusCreated)
	if r.Header.Get(HeaderAccept) == HeaderAppYaml {
		w.Header().Set(HeaderContentType, HeaderAppYaml)
		err := yaml.NewEncoder(w).Encode(res)
		if err != nil {
			log.Errorf("failed to yaml encode 'changelog' report with '%v'", err)
		}
		return
	}
	w.Header().Set(HeaderContentType, HeaderAppJson)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Errorf("failed to json encode 'changelog' report with '%v'", err)
	}
}

func ChangelogFromFile(w http.ResponseWriter, r *http.Request) {

	dir, base, revision, code := CreateFiles(r)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}
	defer CloseFile(base)
	defer CloseFile(revision)
	defer os.RemoveAll(dir)

	changes, code := calcChangelog(r, base.Name(), revision.Name())
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	res := map[string]checker.Changes{
		"changes": changes}
	w.WriteHeader(http.StatusCreated)
	if r.Header.Get(HeaderAccept) == HeaderAppYaml {
		w.Header().Set(HeaderContentType, HeaderAppYaml)
		err := yaml.NewEncoder(w).Encode(res)
		if err != nil {
			log.Errorf("failed to yaml encode 'changelog' report with '%v'", err)
		}
	} else {
		w.Header().Set(HeaderContentType, HeaderAppJson)
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Errorf("failed to json encode 'changelog' report with '%v'", err)
		}
	}
}

func calcChangelog(r *http.Request, base string, revision string) (checker.Changes, int) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	s1, err := load.LoadSpecInfo(loader, base)
	if err != nil {
		log.Infof("failed to load base spec from %q with %v", base, err)
		return nil, http.StatusBadRequest
	}
	s2, err := load.LoadSpecInfo(loader, revision)
	if err != nil {
		log.Infof("failed to load revision spec from %q with %v", revision, err)
		return nil, http.StatusBadRequest
	}

	diffReport, operationsSources, err := diff.GetWithOperationsSourcesMap(
		CreateConfig(r).WithCheckBreaking(), s1, s2)
	if err != nil {
		log.Errorf("failed to 'diff.GetWithOperationsSourcesMap' with %v", err)
		return nil, http.StatusInternalServerError
	}

	c := checker.GetChecks([]string{})
	c.Localizer = *localizations.New(getLocal(r), "en")

	return checker.CheckBackwardCompatibilityUntilLevel(c, diffReport, operationsSources, checker.INFO), http.StatusOK
}

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
	"gopkg.in/yaml.v3"
)

func BreakingChangesFromUri(w http.ResponseWriter, r *http.Request) {

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

	calcBreakingChanges(w, r, base, revision)
}

func BreakingChangesFromFile(w http.ResponseWriter, r *http.Request) {

	dir, base, revision, code := CreateFiles(r)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}
	defer CloseFile(base)
	defer CloseFile(revision)
	defer os.RemoveAll(dir)

	calcBreakingChanges(w, r, base.Name(), revision.Name())
}

func calcBreakingChanges(w http.ResponseWriter, r *http.Request, base string, revision string) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	s1, err := checker.LoadOpenAPISpecInfo(loader, base)
	if err != nil {
		log.Infof("failed to load base spec from %q with %v", base, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s2, err := checker.LoadOpenAPISpecInfo(loader, revision)
	if err != nil {
		log.Infof("failed to load revision spec from %q with %v", revision, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	diffReport, operationsSources, err := diff.GetWithOperationsSourcesMap(
		CreateConfig(r).WithCheckBreaking(), s1, s2)
	if err != nil {
		log.Errorf("failed to 'diff.GetWithOperationsSourcesMap' with %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := checker.GetDefaultChecks()
	c.Localizer = *localizations.New(getLocal(r), "en")

	res := map[string][]checker.BackwardCompatibilityError{
		"breaking-changes": checker.CheckBackwardCompatibility(c, diffReport, operationsSources)}
	w.WriteHeader(http.StatusCreated)
	if r.Header.Get(HeaderAccept) == HeaderAppYaml {
		w.Header().Set(HeaderContentType, HeaderAppYaml)
		err := yaml.NewEncoder(w).Encode(res)
		if err != nil {
			log.Errorf("failed to yaml encode 'breaking-changes' report with '%v'", err)
		}
	} else {
		w.Header().Set(HeaderContentType, HeaderAppJson)
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Errorf("failed to json encode 'breaking-changes' report with '%v'", err)
		}
	}
}

func getLocal(r *http.Request) string {

	local := r.URL.Query()["local"]
	if local != nil && local[0] != "" {
		return local[0]
	}

	return "en"
}

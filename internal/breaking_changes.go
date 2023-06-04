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

func BreakingChanges(w http.ResponseWriter, r *http.Request) {

	dir, base, revision, code := CreateFiles(r)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}
	defer CloseFile(base)
	defer CloseFile(revision)
	defer os.RemoveAll(dir)

	breakingChanges, code := calcBreakingChanges(r, base.Name(), revision.Name())
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	res := map[string][]checker.BackwardCompatibilityError{"breaking-changes": breakingChanges}
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

func calcBreakingChanges(r *http.Request, base string, revision string) ([]checker.BackwardCompatibilityError, int) {

	config := CreateConfig(r).WithCheckBreaking()

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	s1, err := checker.LoadOpenAPISpecInfo(loader, base)
	if err != nil {
		log.Infof("failed to load base spec from %q with %v", base, err)
		return nil, http.StatusBadRequest
	}
	s2, err := checker.LoadOpenAPISpecInfo(loader, revision)
	if err != nil {
		log.Infof("failed to load revision spec from %q with %v", revision, err)
		return nil, http.StatusBadRequest
	}
	diffReport, operationsSources, err := diff.GetWithOperationsSourcesMap(config, s1, s2)
	if err != nil {
		log.Errorf("failed to 'diff.GetWithOperationsSourcesMap' with %v", err)
		return nil, http.StatusInternalServerError
	}

	c := checker.GetDefaultChecks()
	c.Localizer = *localizations.New(getLocal(r), "en")

	return checker.CheckBackwardCompatibility(c, diffReport, operationsSources), http.StatusOK
}

func getLocal(r *http.Request) string {

	local := r.URL.Query()["local"]
	if local != nil && local[0] != "" {
		return local[0]
	}

	return "en"
}

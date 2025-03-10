package internal

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"gopkg.in/yaml.v3"
)

func (h *Handler) ChangelogFromUri(w http.ResponseWriter, r *http.Request) {

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

	acceptHeader := GetAcceptHeader(r)
	_ = h.SendTelemetry(r, CommandChangelog, []string{base, revision}, acceptHeader)

	changes, code := calcChangelog(r, base, revision)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	res := map[string]checker.Changes{"changelog": changes}
	w.WriteHeader(http.StatusCreated)
	if acceptHeader == HeaderAppYaml {
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

func (h *Handler) ChangelogFromFile(w http.ResponseWriter, r *http.Request) {

	dir, base, revision, code := CreateFiles(r)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}
	defer CloseFile(base)
	defer CloseFile(revision)
	defer os.RemoveAll(dir)

	acceptHeader := GetAcceptHeader(r)
	_ = h.SendTelemetry(r, CommandChangelog, []string{base.Name(), revision.Name()}, acceptHeader)

	changes, code := calcChangelog(r, base.Name(), revision.Name())
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	res := map[string]formatters.Changes{"changes": formatters.NewChanges(changes, checker.NewDefaultLocalizer())}
	w.WriteHeader(http.StatusCreated)
	if acceptHeader == HeaderAppYaml {
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

	s1, err := load.NewSpecInfo(loader, load.NewSource(base))
	if err != nil {
		log.Infof("failed to load base spec from %q with %v", base, err)
		return nil, http.StatusBadRequest
	}
	s2, err := load.NewSpecInfo(loader, load.NewSource(revision))
	if err != nil {
		log.Infof("failed to load revision spec from %q with %v", revision, err)
		return nil, http.StatusBadRequest
	}

	diffReport, operationsSources, err := diff.GetWithOperationsSourcesMap(
		CreateConfig(r), s1, s2)
	if err != nil {
		log.Errorf("failed to 'diff.GetWithOperationsSourcesMap' with %v", err)
		return nil, http.StatusInternalServerError
	}

	return checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), diffReport, operationsSources, checker.INFO), http.StatusOK
}

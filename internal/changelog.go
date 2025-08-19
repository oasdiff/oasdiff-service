package internal

import (
	"fmt"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	log "github.com/sirupsen/logrus"
)

const CHANGELOG_LEVEL = checker.INFO

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

	getChangelog(w, r, base, revision, CHANGELOG_LEVEL)
}

func (h *Handler) ChangelogFromFile(w http.ResponseWriter, r *http.Request) {

	dir, base, revision, err := CreateFiles(r)
	if err != nil {
		log.Errorf("failed to create files with %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer CloseFile(base)
	defer CloseFile(revision)
	defer os.RemoveAll(dir)

	getChangelog(w, r, base.Name(), revision.Name(), CHANGELOG_LEVEL)
}

func getChangelog(w http.ResponseWriter, r *http.Request, base string, revision string, level checker.Level) {
	specInfoPair, err := getSpecInfoPair(base, revision)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	changes, err := calcChangelog(r, specInfoPair, level)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	contentType := getContentType(GetAcceptHeader(r))
	languageCode := GetLanguageCode(GetAcceptLanguageHeader(r))

	out, err := getChangelogOutput(changes, contentType, specInfoPair, languageCode)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set(HeaderContentType, contentType)
	_, _ = w.Write(out)
}

func getContentType(acceptHeader string) string {
	if acceptHeader == "" || acceptHeader == "*/*" {
		return HeaderAppJson
	}
	return acceptHeader
}

func getChangelogOutput(changes checker.Changes, contentType string, specInfoPair *load.SpecInfoPair, languageCode string) ([]byte, error) {
	switch contentType {
	case HeaderAppYaml:
		out, err := formatters.YAMLFormatter{
			Localizer: checker.NewLocalizer(languageCode),
		}.RenderChangelog(changes, formatters.RenderOpts{WrapInObject: true}, specInfoPair.GetBaseVersion(), specInfoPair.GetRevisionVersion())
		if err != nil {
			return nil, fmt.Errorf("failed to yaml encode 'breaking-changes' report with '%v'", err)
		}
		return out, nil
	case HeaderAppJson:
		out, err := formatters.JSONFormatter{
			Localizer: checker.NewLocalizer(languageCode),
		}.RenderChangelog(changes, formatters.RenderOpts{WrapInObject: true}, specInfoPair.GetBaseVersion(), specInfoPair.GetRevisionVersion())
		if err != nil {
			return nil, fmt.Errorf("failed to json encode 'breaking-changes' report with '%v'", err)
		}
		return out, nil
	case HeaderTextHtml:
		out, err := formatters.HTMLFormatter{
			Localizer: checker.NewLocalizer(languageCode),
		}.RenderChangelog(changes, formatters.NewRenderOpts(), specInfoPair.GetBaseVersion(), specInfoPair.GetRevisionVersion())
		if err != nil {
			return nil, fmt.Errorf("failed to html encode 'breaking-changes' report with '%v'", err)
		}
		return out, nil
	case HeaderTextPlain:
		out, err := formatters.TEXTFormatter{
			Localizer: checker.NewLocalizer(languageCode),
		}.RenderChangelog(changes, formatters.NewRenderOpts(), specInfoPair.GetBaseVersion(), specInfoPair.GetRevisionVersion())
		if err != nil {
			return nil, fmt.Errorf("failed to text encode 'breaking-changes' report with '%v'", err)
		}
		return out, nil
	case HeaderTextMarkdown:
		out, err := formatters.MarkupFormatter{
			Localizer: checker.NewLocalizer(languageCode),
		}.RenderChangelog(changes, formatters.NewRenderOpts(), specInfoPair.GetBaseVersion(), specInfoPair.GetRevisionVersion())
		if err != nil {
			return nil, fmt.Errorf("failed to markdown encode 'breaking-changes' report with '%v'", err)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported content type '%v'", contentType)
	}
}

func getSpecInfoPair(base string, revision string) (*load.SpecInfoPair, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	s1, err := load.NewSpecInfo(loader, load.NewSource(base))
	if err != nil {
		return nil, fmt.Errorf("failed to load base spec from %q with %v", base, err)
	}
	s2, err := load.NewSpecInfo(loader, load.NewSource(revision))
	if err != nil {
		return nil, fmt.Errorf("failed to load revision spec from %q with %v", revision, err)
	}

	return load.NewSpecInfoPair(s1, s2), nil
}

func calcChangelog(r *http.Request, specInfoPair *load.SpecInfoPair, level checker.Level) (checker.Changes, error) {

	diffReport, operationsSources, err := diff.GetWithOperationsSourcesMap(
		CreateConfig(r), specInfoPair.Base, specInfoPair.Revision)
	if err != nil {
		return nil, fmt.Errorf("failed to 'diff.GetWithOperationsSourcesMap' with %v", err)
	}

	return checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), diffReport, operationsSources, level), nil
}

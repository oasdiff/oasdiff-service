package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	log "github.com/sirupsen/logrus"
)

func (h *Handler) DiffFromUri(w http.ResponseWriter, r *http.Request) {

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

	contentType := getContentType(GetAcceptHeader(r))
	languageCode := GetLanguageCode(GetAcceptLanguageHeader(r))

	diffReport, code := createDiffReport(r, baseSpec, revisionSpec, contentType)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	out, err := getDiffOutput(diffReport, contentType, languageCode)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set(HeaderContentType, contentType)
	_, _ = w.Write(out)
}

func (h *Handler) DiffFromFile(w http.ResponseWriter, r *http.Request) {

	dir, base, revision, err := CreateFiles(r)
	if err != nil {
		log.Errorf("failed to create files with %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer CloseFile(base)
	defer CloseFile(revision)
	defer os.RemoveAll(dir)

	baseSpec, revisionSpec, err := createSpecFromFile(base, revision)
	if err != nil {
		log.Errorf("failed to create spec from files with %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	contentType := getContentType(GetAcceptHeader(r))
	languageCode := GetLanguageCode(GetAcceptLanguageHeader(r))

	diffReport, code := createDiffReport(r, baseSpec, revisionSpec, contentType)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	out, err := getDiffOutput(diffReport, contentType, languageCode)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set(HeaderContentType, contentType)
	_, _ = w.Write(out)
}

func getDiffOutput(diffReport *diff.Diff, contentType string, languageCode string) ([]byte, error) {
	switch contentType {
	case HeaderAppYaml:
		out, err := formatters.YAMLFormatter{
			Localizer: checker.NewLocalizer(languageCode),
		}.RenderDiff(diffReport, formatters.NewRenderOpts())
		if err != nil {
			return nil, fmt.Errorf("failed to yaml encode diff with '%v'", err)
		}
		return out, nil
	case HeaderAppJson:
		out, err := formatters.JSONFormatter{
			Localizer: checker.NewLocalizer(languageCode),
		}.RenderDiff(diffReport, formatters.NewRenderOpts())
		if err != nil {
			return nil, fmt.Errorf("failed to json encode diff with '%v'", err)
		}
		return out, nil
	case HeaderTextHtml:
		out, err := formatters.HTMLFormatter{
			Localizer: checker.NewLocalizer(languageCode),
		}.RenderDiff(diffReport, formatters.NewRenderOpts())
		if err != nil {
			return nil, fmt.Errorf("failed to html encode diff with '%v'", err)
		}
		return out, nil
	case HeaderTextPlain:
		out, err := formatters.TEXTFormatter{
			Localizer: checker.NewLocalizer(languageCode),
		}.RenderDiff(diffReport, formatters.NewRenderOpts())
		if err != nil {
			return nil, fmt.Errorf("failed to text encode diff with '%v'", err)
		}
		return out, nil
	case HeaderTextMarkdown:
		out, err := formatters.MarkupFormatter{
			Localizer: checker.NewLocalizer(languageCode),
		}.RenderDiff(diffReport, formatters.NewRenderOpts())
		if err != nil {
			return nil, fmt.Errorf("failed to markdown encode diff with '%v'", err)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported content type '%v'", contentType)
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

func createSpecFromFile(base *os.File, revision *os.File) (*openapi3.T, *openapi3.T, error) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false

	s1, err := load.NewSpecInfo(loader, load.NewSource(base.Name()))
	if err != nil {

		return nil, nil, fmt.Errorf("failed to load base spec from '%s' with '%v'", base.Name(), err)
	}

	s2, err := load.NewSpecInfo(loader, load.NewSource(revision.Name()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load revision spec from '%s' with '%v'", revision.Name(), err)
	}

	return s1.Spec, s2.Spec, nil
}

func createDiffReport(r *http.Request, s1 *openapi3.T, s2 *openapi3.T, contentType string) (*diff.Diff, int) {

	config := CreateConfig(r)

	// exclude endpoints in json output
	if contentType == HeaderAppJson {
		config.ExcludeElements.Add(diff.ExcludeEndpointsOption)
	}

	diffReport, err := diff.Get(config, s1, s2)
	if err != nil {
		log.Infof("failed to calculate diff between a pair of OpenAPI objects '%s' with %v", s1.Info.Title, err)
		return nil, http.StatusBadRequest
	}

	return diffReport, http.StatusOK
}

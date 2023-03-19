package internal

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tufin/oasdiff/diff"
)

func CreateConfig(r *http.Request) *diff.Config {

	config := diff.NewConfig()
	config.ExcludeExamples = getBoolQueryString(r, "exclude-examples", config.ExcludeExamples)
	config.ExcludeDescription = getBoolQueryString(r, "exclude-description", config.ExcludeDescription)
	config.PathFilter = getQueryString(r, "path-filter", config.PathFilter)
	config.FilterExtension = getQueryString(r, "filter-extension", config.FilterExtension)
	config.PathPrefixBase = getQueryString(r, "path-prefix-base", config.PathPrefixBase)
	config.PathPrefixRevision = getQueryString(r, "path-prefix-revision", config.PathPrefixRevision)
	config.PathStripPrefixBase = getQueryString(r, "path-strip-prefix-base", config.PathStripPrefixBase)
	config.PathStripPrefixRevision = getQueryString(r, "path-strip-prefix-revision", config.PathStripPrefixRevision)
	config.BreakingOnly = false // breaking-only is deprecated
	config.DeprecationDays = getIntQueryString(r, "deprecation-days", config.DeprecationDays)
	config.ExcludeEndpoints = getBoolQueryString(r, "exclude-endpoints", config.ExcludeEndpoints)
	// config.IncludeExtensions = StringSet{}

	return config
}

func getQueryString(r *http.Request, key string, defaultValue string) string {

	if val, ok := r.URL.Query()[key]; ok {
		return val[0]
	}

	return defaultValue
}

func getIntQueryString(r *http.Request, key string, defaultValue int) int {

	if val, ok := r.URL.Query()[key]; ok {
		if res, err := strconv.Atoi(val[0]); err == nil && res >= 0 {
			return res
		}
		log.Infof("invalid query string '%s: %s' (using default '%d')", key, val[0], defaultValue)
	}

	return defaultValue
}

func getBoolQueryString(r *http.Request, key string, defaultValue bool) bool {

	if val, ok := r.URL.Query()[key]; ok {
		if val[0] == "true" {
			return true
		}
		if val[0] == "false" {
			return false
		}
		log.Infof("invalid query string '%s: %s' (using default '%v')", key, val[0], defaultValue)
	}

	return defaultValue
}

func CreateFiles(r *http.Request) (string, *os.File, *os.File, int) {

	// create a temporary directory
	dir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		log.Errorf("failed to make temp dir with %v", err)
		return "", nil, nil, http.StatusInternalServerError
	}

	// create temporary files for base and revision
	base, code := createTempFile(dir, "base")
	if code != http.StatusOK {
		os.RemoveAll(dir)
		return "", nil, nil, code
	}
	revision, code := createTempFile(dir, "revision")
	if code != http.StatusOK {
		os.RemoveAll(dir)
		CloseFile(base)
		return "", nil, nil, code
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// 32 MB is the default used by FormFile() function
		if err := r.ParseMultipartForm(4); err != nil {
			log.Infof("failed to parse 'multipart/form-data' request files with '%v'", err)
			return "", nil, nil, http.StatusBadRequest
		}
		if code := copyMultipartFormData(r, "base", base); code != http.StatusOK {
			return "", nil, nil, code
		}
		if code := copyMultipartFormData(r, "revision", revision); code != http.StatusOK {
			return "", nil, nil, code
		}
	} else if contentType == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			log.Errorf("failed to parse form request with '%v'", err)
			return "", nil, nil, http.StatusBadRequest
		}
		if code := copyFormData(r, "base", base); code != http.StatusOK {
			return "", nil, nil, code
		}
		if code := copyFormData(r, "revision", revision); code != http.StatusOK {
			return "", nil, nil, code
		}
	} else {
		log.Infof("unsupported content type '%s'", contentType)
		return "", nil, nil, http.StatusBadRequest
	}

	return dir, base, revision, http.StatusOK
}

func createTempFile(dir string, filename string) (*os.File, int) {

	res, err := os.CreateTemp(dir, "")
	if err != nil {
		log.Errorf("failed to create temp file '%s' with '%v'", filename, err)
		return nil, http.StatusInternalServerError
	}

	return res, http.StatusOK
}

func copyMultipartFormData(r *http.Request, filename string, res *os.File) int {

	// a reference to the fileHeaders are accessible only after ParseMultipartForm is called
	files := r.MultipartForm.File[filename]
	for _, fileHeader := range files {
		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			log.Errorf("failed to create temp file with %v", err)
			return http.StatusInternalServerError
		}
		defer file.Close()

		_, err = io.Copy(res, file)
		if err != nil {
			log.Errorf("failed to copy file %q from HTTP request with %v", fileHeader.Filename, err)
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

func copyFormData(r *http.Request, filename string, res *os.File) int {

	data := r.FormValue(filename)
	if data == "" {
		log.Infof("bad request: empty spec '%s'", filename)
		return http.StatusBadRequest
	}

	_, err := io.Copy(res, strings.NewReader(data))
	if err != nil {
		log.Errorf("failed to copy form value '%s' from HTTP request with %v", filename, err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func CloseFile(f *os.File) {

	err := f.Close()
	if err != nil {
		log.Errorf("failed to close file with %v", err)
	}
}

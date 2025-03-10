package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/oasdiff/oasdiff/diff"
)

func CreateConfig(r *http.Request) *diff.Config {

	config := diff.NewConfig()
	config.MatchPath = GetQueryString(r, "path-filter", config.MatchPath)
	config.FilterExtension = GetQueryString(r, "filter-extension", config.FilterExtension)
	config.PathPrefixBase = GetQueryString(r, "path-prefix-base", config.PathPrefixBase)
	config.PathPrefixRevision = GetQueryString(r, "path-prefix-revision", config.PathPrefixRevision)
	config.PathStripPrefixBase = GetQueryString(r, "path-strip-prefix-base", config.PathStripPrefixBase)
	config.PathStripPrefixRevision = GetQueryString(r, "path-strip-prefix-revision", config.PathStripPrefixRevision)

	return config
}

func CreateFiles(r *http.Request) (string, *os.File, *os.File, int) {

	// create a temporary directory
	dir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		log.Errorf("failed to make temp dir with %v", err)
		return "", nil, nil, http.StatusInternalServerError
	}

	// create temporary files for base and revision
	base, code := createFile(dir, "base")
	if code != http.StatusOK {
		os.RemoveAll(dir)
		return "", nil, nil, code
	}
	revision, code := createFile(dir, "revision")
	if code != http.StatusOK {
		os.RemoveAll(dir)
		CloseFile(base)
		return "", nil, nil, code
	}

	contentType := r.Header.Get(HeaderContentType)
	if strings.HasPrefix(contentType, HeaderMultipartFormData) {
		// 32 MB is the default used by FormFile() function
		if err := r.ParseMultipartForm(4); err != nil {
			log.Infof("failed to parse '%s' request files with '%v'", HeaderMultipartFormData, err)
			return "", nil, nil, http.StatusBadRequest
		}
		if code := copyMultipartFormData(r, "base", base); code != http.StatusOK {
			return "", nil, nil, code
		}
		if code := copyMultipartFormData(r, "revision", revision); code != http.StatusOK {
			return "", nil, nil, code
		}
	} else if contentType == HeaderAppFormUrlEncoded {
		if err := r.ParseForm(); err != nil {
			log.Infof("failed to parse '%s' request with '%v'", HeaderAppFormUrlEncoded, err)
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

func CloseFile(f *os.File) {

	err := f.Close()
	if err != nil {
		log.Errorf("failed to close file with %v", err)
	}
}

func createFile(dir string, filename string) (*os.File, int) {

	f := fmt.Sprintf("%s/%s", dir, filename)
	res, err := os.Create(f)
	if err != nil {
		log.Errorf("failed to create file '%s' with '%v'", f, err)
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

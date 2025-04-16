package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/oasdiff/oasdiff/diff"
	log "github.com/sirupsen/logrus"
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

func CreateFiles(r *http.Request) (string, *os.File, *os.File, error) {

	// create a temporary directory
	dir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to make temp dir with %v", err)
	}

	// create temporary files for base and revision
	base, err := createFile(dir, "base")
	if err != nil {
		os.RemoveAll(dir)
		return "", nil, nil, err
	}
	revision, err := createFile(dir, "revision")
	if err != nil {
		os.RemoveAll(dir)
		CloseFile(base)
		return "", nil, nil, err
	}

	contentType := r.Header.Get(HeaderContentType)
	if strings.HasPrefix(contentType, HeaderMultipartFormData) {
		// 32 MB is the default used by FormFile() function
		if err := r.ParseMultipartForm(4); err != nil {
			return "", nil, nil, fmt.Errorf("failed to parse '%s' request files with '%v'", HeaderMultipartFormData, err)
		}
		if err := copyMultipartFormData(r, "base", base); err != nil {
			return "", nil, nil, err
		}
		if err := copyMultipartFormData(r, "revision", revision); err != nil {
			return "", nil, nil, err
		}
	} else if contentType == HeaderAppFormUrlEncoded {
		if err := r.ParseForm(); err != nil {
			return "", nil, nil, fmt.Errorf("failed to parse '%s' request with '%v'", HeaderAppFormUrlEncoded, err)
		}
		if err := copyFormData(r, "base", base); err != nil {
			return "", nil, nil, err
		}
		if err := copyFormData(r, "revision", revision); err != nil {
			return "", nil, nil, err
		}
	} else {
		return "", nil, nil, fmt.Errorf("unsupported content type '%s'", contentType)
	}

	return dir, base, revision, nil
}

func CloseFile(f *os.File) {

	err := f.Close()
	if err != nil {
		log.Errorf("failed to close file with %v", err)
	}
}

func createFile(dir string, filename string) (*os.File, error) {

	f := fmt.Sprintf("%s/%s", dir, filename)
	res, err := os.Create(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create file '%s' with '%v'", f, err)
	}

	return res, nil
}

func copyMultipartFormData(r *http.Request, filename string, res *os.File) error {

	// a reference to the fileHeaders are accessible only after ParseMultipartForm is called
	files := r.MultipartForm.File[filename]
	for _, fileHeader := range files {
		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("failed to create temp file with %v", err)
		}
		defer file.Close()

		_, err = io.Copy(res, file)
		if err != nil {
			return fmt.Errorf("failed to copy file %q from HTTP request with %v", fileHeader.Filename, err)
		}
	}

	return nil
}

func copyFormData(r *http.Request, filename string, res *os.File) error {

	data := r.FormValue(filename)
	if data == "" {
		return fmt.Errorf("bad request: empty spec '%s'", filename)
	}

	_, err := io.Copy(res, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to copy form value '%s' from HTTP request with %v", filename, err)
	}

	return nil
}

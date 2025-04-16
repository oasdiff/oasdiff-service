package internal

import "net/http"

const (
	HeaderContentType       = "Content-Type"
	HeaderAccept            = "Accept"
	HeaderAppYaml           = "application/yaml"
	HeaderAppJson           = "application/json"
	HeaderTextHtml          = "text/html"
	HeaderTextPlain         = "text/plain"
	HeaderMultipartFormData = "multipart/form-data"
	HeaderAppFormUrlEncoded = "application/x-www-form-urlencoded"
)

func GetAcceptHeader(r *http.Request) string {

	return r.Header.Get(HeaderAccept)
}

func GetQueryString(r *http.Request, key string, defaultValue string) string {

	if val, ok := r.URL.Query()[key]; ok {
		return val[0]
	}

	return defaultValue
}

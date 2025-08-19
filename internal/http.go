package internal

import (
	"net/http"
	"strings"
)

const (
	HeaderContentType       = "Content-Type"
	HeaderAccept            = "Accept"
	HeaderAcceptLanguage    = "Accept-Language"
	HeaderAppYaml           = "application/yaml"
	HeaderAppJson           = "application/json"
	HeaderTextHtml          = "text/html"
	HeaderTextPlain         = "text/plain"
	HeaderTextMarkdown      = "text/markdown"
	HeaderMultipartFormData = "multipart/form-data"
	HeaderAppFormUrlEncoded = "application/x-www-form-urlencoded"
)

func GetAcceptHeader(r *http.Request) string {

	return r.Header.Get(HeaderAccept)
}

func GetAcceptLanguageHeader(r *http.Request) string {

	return r.Header.Get(HeaderAcceptLanguage)
}

func GetLanguageCode(acceptLanguageHeader string) string {
	if acceptLanguageHeader == "" {
		return "en"
	}

	languages := strings.Split(acceptLanguageHeader, ",")
	for _, lang := range languages {
		lang = strings.TrimSpace(lang)
		lang = strings.Split(lang, ";")[0]

		switch strings.ToLower(lang) {
		case "en", "en-us", "en-gb":
			return "en"
		case "ru", "ru-ru":
			return "ru"
		case "pt-br", "pt":
			return "pt-br"
		case "es", "es-es", "es-mx":
			return "es"
		}
	}

	return "en"
}

func GetQueryString(r *http.Request, key string, defaultValue string) string {

	if val, ok := r.URL.Query()[key]; ok {
		return val[0]
	}

	return defaultValue
}

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/oasdiff/go-common/ds"
	"github.com/oasdiff/go-common/env"
	"github.com/oasdiff/go-common/tenant"
	"github.com/oasdiff/oasdiff-service/internal"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

func main() {

	var (
		diff            = fmt.Sprintf("/tenants/{%s}/diff", tenant.PathParamTenantId)
		breakingChanges = fmt.Sprintf("/tenants/{%s}/breaking-changes", tenant.PathParamTenantId)
		changelog       = fmt.Sprintf("/tenants/{%s}/changelog", tenant.PathParamTenantId)

		v = tenant.NewValidator(ds.NewClient(env.GetGCPProject(), env.GetGCPDatastoreNamespace()))
		h = internal.NewHandler()
	)

	serve(
		[]string{
			fmt.Sprintf("/tenants/{%s}/docs.html", tenant.PathParamTenantId),
			fmt.Sprintf("/tenants/{%s}/openapi.yaml", tenant.PathParamTenantId),
			diff, diff, diff,
			breakingChanges, breakingChanges, breakingChanges,
			changelog, changelog, changelog,
		},
		[]string{
			http.MethodGet,
			http.MethodGet,
			http.MethodPost, http.MethodGet, http.MethodOptions,
			http.MethodPost, http.MethodGet, http.MethodOptions,
			http.MethodPost, http.MethodGet, http.MethodOptions,
		},
		[]func(http.ResponseWriter, *http.Request){
			func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "/app/docs/docs.html") },
			func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "/app/docs/openapi.yaml") },
			access(h.DiffFromFile), access(h.DiffFromUri), options([]string{http.MethodPost, http.MethodGet}),
			access(h.BreakingChangesFromFile), access(h.BreakingChangesFromUri), options([]string{http.MethodPost, http.MethodGet}),
			access(h.ChangelogFromFile), access(h.ChangelogFromUri), options([]string{http.MethodPost, http.MethodGet}),
		},
		v.Validate,
	)
}

func access(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next(w, r)
	}
}

func options(methods []string) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
	}
}

func serve(path []string, method []string,
	handle []func(http.ResponseWriter, *http.Request), mwf ...mux.MiddlewareFunc) {

	initLogger()
	logVersion()

	router := mux.NewRouter()
	router.Use(mwf...)
	for i := 0; i < len(path); i++ {
		router.HandleFunc(path[i], handle[i]).Methods(method[i])
	}
	server := &http.Server{
		Addr: fmt.Sprintf("%s:%s", "0.0.0.0", "8080"),
		// avoid slowloris attacks
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Error(err)
		os.Exit(0)
	}
}

func logVersion() {

	log.Infof("%s/%s, %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func initLogger() {

	// log.SetReportCaller(true)
	initLoggerOutput()
	log.SetLevel(getLogLevel())
}

func initLoggerOutput() {

	log.SetOutput(io.Discard) // Send all logs to nowhere by default - this is required to avoid duplicate log messages
	log.AddHook(filename.NewHook())
	log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
		},
	})
	log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.WarnLevel,
			log.InfoLevel,
			log.DebugLevel,
			log.TraceLevel,
		},
	})
}

func getLogLevel() log.Level {

	level := os.Getenv("LOG_LEVEL")
	if strings.EqualFold(level, "debug") {
		return log.DebugLevel
	} else if strings.EqualFold(level, "warn") {
		return log.WarnLevel
	} else if strings.EqualFold(level, "error") {
		return log.ErrorLevel
	}
	return log.InfoLevel
}

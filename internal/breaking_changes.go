package internal

import (
	"net/http"
	"os"

	"github.com/oasdiff/oasdiff/checker"
	log "github.com/sirupsen/logrus"
)

const BREAKING_LEVEL = checker.WARN

func (h *Handler) BreakingChangesFromUri(w http.ResponseWriter, r *http.Request) {

	base := GetQueryString(r, "base", "")
	if base == "" {
		log.Error("no base url provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	revision := GetQueryString(r, "revision", "")
	if revision == "" {
		log.Error("no revision url provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	getChangelog(w, r, base, revision, BREAKING_LEVEL)
}

func (h *Handler) BreakingChangesFromFile(w http.ResponseWriter, r *http.Request) {

	dir, base, revision, err := CreateFiles(r)
	if err != nil {
		log.Errorf("failed to create files with %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer CloseFile(base)
	defer CloseFile(revision)
	defer os.RemoveAll(dir)

	getChangelog(w, r, base.Name(), revision.Name(), BREAKING_LEVEL)
}

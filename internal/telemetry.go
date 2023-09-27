package internal

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/oasdiff/go-common/tenant"
	"github.com/oasdiff/telemetry/model"
	"github.com/sirupsen/logrus"
)

const (
	CommandDiff      = "diff"
	CommandChangelog = "changelog"
	CommandBreaking  = "breaking"

	version = "v1.8.0"
)

func (h *Handler) SendTelemetry(r *http.Request, platform string, cmd string, args []string, acceptHeader string) error {

	app := fmt.Sprintf("%s-service", model.Application)
	t := model.NewTelemetry(app, version, cmd, args, toFlags(acceptHeader), mux.Vars(r)[tenant.PathParamTenantId], platform)
	if err := h.collector.Send(t); err != nil {
		logrus.Errorf("failed to send telemetry '%+v' with '%v'", t, err)
		return err
	}

	return nil
}

func toFlags(acceptHeader string) map[string]string {

	if acceptHeader == HeaderAppYaml {
		return map[string]string{"format": "yaml"}
	}

	return map[string]string{}
}

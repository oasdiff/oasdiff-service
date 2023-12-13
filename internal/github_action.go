package internal

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/tufin/oasdiff/checker"
	"github.com/tufin/oasdiff/formatters"
)

type githubActionResponse struct {
	Summary     string
	Annotations []byte
}

func GetGitHubActionResponse(changes checker.Changes) (*githubActionResponse, error) {

	gitHubFormatter, err := formatters.Lookup(string(formatters.FormatGithubActions), formatters.DefaultFormatterOpts())
	if err != nil {
		slog.Error("failed to lookup for GitHub formatter", "error", err)
		return nil, err
	}

	annotations, err := gitHubFormatter.RenderBreakingChanges(changes, formatters.NewRenderOpts())
	if err != nil {
		slog.Error("failed to 'RenderChangelog' for GitHub formatter", "error", err)
		return nil, err
	}

	return &githubActionResponse{
		Summary:     getChangesTitle(changes),
		Annotations: annotations,
	}, nil
}

func getChangesTitle(changes checker.Changes) string {

	count := changes.GetLevelCount()
	return fmt.Sprintf("%d changes: %d error, %d warning, %d info",
		len(changes), count[checker.ERR], count[checker.WARN], count[checker.INFO])
}

func getGitHubAnnotations(changes checker.Changes) string {

	var buffer bytes.Buffer
	for _, currChange := range changes {
		buffer.WriteString(toAnnotation(currChange))
	}

	return buffer.String()
}

func toAnnotation(currChange checker.Change) string {

	return getAnnotationLevel()
}

func getAnnotationLevel() string {
	panic("unimplemented")
}

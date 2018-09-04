package linters

import (
	"strings"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

type teamLinter struct{}

func NewTeamLinter() teamLinter {
	return teamLinter{}
}

func (teamLinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "Manifest"
	result.DocsURL = "https://docs.halfpipe.io/manifest/"

	if manifest.Team == "" {
		result.AddError(errors.NewMissingField("team"))
	} else if strings.ToLower(manifest.Team) != manifest.Team {
		result.AddWarning(errors.NewInvalidField("team", "team should be lower case"))
	}

	if manifest.Pipeline == "" {
		result.AddError(errors.NewMissingField("pipeline"))
	}

	if strings.Contains(manifest.Pipeline, " ") {
		result.AddError(errors.NewInvalidField("pipeline", "pipeline name must not contains spaces!"))
	}

	return
}

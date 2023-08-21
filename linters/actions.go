package linters

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
)

type ActionsLinter struct{}

func NewActionsLinter() Linter {
	return ActionsLinter{}
}

func (linter ActionsLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "GitHub Actions"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/github-actions/overview/"
	if man.Platform.IsActions() {
		result.Add(unsupportedTasks(man.Tasks, man, "tasks")...)
		result.Add(unsupportedTriggers(man.Triggers)...)
		result.Add(unsupportedFeatures(man.FeatureToggles)...)
	}
	return result
}

func unsupportedTasks(tasks manifest.TaskList, man manifest.Manifest, taskListId string) (errors []error) {
	for i, t := range tasks {
		taskIdx := fmt.Sprintf("%s[%v]", taskListId, i)

		appendError := func(err error) {
			errors = append(errors, fmt.Errorf("%s %w", taskIdx, err))
		}

		switch task := t.(type) {
		case manifest.Parallel:
			errors = append(errors, unsupportedTasks(task.Tasks, man, taskIdx)...)
		case manifest.Sequence:
			errors = append(errors, unsupportedTasks(task.Tasks, man, taskIdx)...)
		default:
			if task.IsManualTrigger() {
				appendError(ErrUnsupportedManualTrigger.AsWarning())
			}
		}

		switch task := t.(type) {
		case manifest.DeployCF:
			if task.Rolling {
				appendError(ErrUnsupportedRolling.AsWarning())
			}
		case manifest.DockerPush:
			for _, trigger := range man.Triggers {
				if t, ok := trigger.(manifest.DockerTrigger); ok {
					if t.Image == task.Image {
						appendError(ErrDockerTriggerLoop.WithValue(t.Image).AsWarning())
					}
				}
			}
		case manifest.ConsumerIntegrationTest:
			if task.UseCovenant {
				appendError(ErrUnsupportedCovenant.AsWarning())
			}
		}

	}
	return errors
}

func unsupportedTriggers(triggers manifest.TriggerList) (errors []error) {
	for i, trigger := range triggers {
		appendError := func(err error) {
			errors = append(errors, fmt.Errorf("triggers[%v] %w", i, err))
		}

		switch t := trigger.(type) {
		case manifest.GitTrigger:
			if t.PrivateKey != "" {
				appendError(ErrUnsupportedGitPrivateKey.AsWarning())
			}
			if t.URI != "" {
				appendError(ErrUnsupportedGitUri.AsWarning())
			}
		case manifest.PipelineTrigger:
			appendError(ErrUnsupportedPipelineTrigger.AsWarning())
		default:
			// ok
		}
	}
	return errors
}

func unsupportedFeatures(features manifest.FeatureToggles) (errors []error) {
	return []error{}
}

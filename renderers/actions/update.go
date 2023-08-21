package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) updateSteps(task manifest.Update, man manifest.Manifest) Steps {

	cdPrefix := ""
	if a.workingDir != "" {
		cdPrefix = fmt.Sprintf("cd %s; ", a.workingDir)
	}

	update := Step{
		Name: "Sync workflow with halfpipe manifest",
		ID:   "sync",
		Uses: "docker://eu.gcr.io/halfpipe-io/halfpipe-auto-update",
		With: With{
			"args":       fmt.Sprintf(`-c "%supdate-actions-workflow"`, cdPrefix),
			"entrypoint": "/bin/bash",
		},
		Env: Env{
			"HALFPIPE_FILE_PATH": a.halfpipeFilePath,
		},
	}

	steps := Steps{update}

	if task.TagRepo {
		tag := man.PipelineName() + "/v$BUILD_VERSION"
		tagStep := Step{
			Name: "Tag commit with " + tag,
			Run:  fmt.Sprintf("git tag -f %s\ngit push origin %s", tag, tag),
		}
		steps = append(steps, tagStep)
	}

	return steps
}

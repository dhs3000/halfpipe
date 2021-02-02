package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func (a *Actions) runSteps(task manifest.Run) Steps {
	steps := Steps{checkoutCode}
	if task.ReadsFromArtifacts() {
		steps = append(steps, a.restoreArtifacts()...)
	}
	run := Step{
		Name: "run",
		Env:  Env(task.Vars),
	}

	prefix := ""
	if a.workingDir != "" {
		prefix = fmt.Sprintf("cd %s;", a.workingDir)
	}
	if task.Docker.Image != "" {
		run.Uses = "docker://" + task.Docker.Image
		run.With = With{
			{"entrypoint", "/bin/sh"},
			{"args", fmt.Sprintf(`-c "%s %s"`, prefix, strings.Replace(task.Script, `"`, `\"`, -1))},
		}
	} else {
		run.Run = task.Script
	}

	steps = append(steps, dockerLogin(task.Docker.Image, task.Docker.Username, task.Docker.Password)...)
	steps = append(steps, run)

	if task.SavesArtifacts() {
		steps = append(steps, a.saveArtifacts(task.SaveArtifacts)...)
	}
	if task.SavesArtifactsOnFailure() {
		steps = append(steps, a.saveArtifactsOnFailure(task.SaveArtifactsOnFailure)...)
	}
	return steps
}

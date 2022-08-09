package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"path"
	"strings"
)

const tagList_Dir = "tagList"

var tagListFile = path.Join(tagList_Dir, "tagList")

func (c Concourse) dockerPushJob(task manifest.DockerPush, basePath string, ociBuild bool, updatePipeline bool) atc.JobConfig {
	var steps []atc.Step
	resourceName := manifest.DockerTrigger{Image: task.Image}.GetTriggerName()

	fullBasePath := path.Join(gitDir, basePath)
	if task.RestoreArtifacts {
		fullBasePath = path.Join(dockerBuildTmpDir, basePath)
	}

	steps = append(steps, restoreArtifacts(task)...)
	steps = append(steps, buildAndPush(task, resourceName, ociBuild, fullBasePath, task.RestoreArtifacts, updatePipeline)...)

	return atc.JobConfig{
		Name:         task.GetName(),
		Serial:       true,
		PlanSequence: steps,
	}
}

func restoreArtifacts(task manifest.DockerPush) []atc.Step {
	if task.RestoreArtifacts {
		copyArtifact := &atc.TaskStep{
			Name: "copying-git-repo-and-artifacts-to-a-temporary-build-dir",
			Config: &atc.TaskConfig{
				Platform: "linux",
				ImageResource: &atc.ImageResource{
					Type: "docker-image",
					Source: atc.Source{
						"repository": "alpine",
					},
				},
				Run: atc.TaskRunConfig{
					Path: "/bin/sh",
					Args: []string{"-c", strings.Join([]string{
						fmt.Sprintf("cp -r %s/. %s", gitDir, dockerBuildTmpDir),
						fmt.Sprintf("cp -r %s/. %s", artifactsInDir, dockerBuildTmpDir),
					}, "\n")},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: gitDir},
					{Name: artifactsName},
				},
				Outputs: []atc.TaskOutputConfig{
					{Name: dockerBuildTmpDir},
				},
			},
		}
		return append([]atc.Step{}, stepWithAttemptsAndTimeout(copyArtifact, task.GetAttempts(), task.Timeout))
	}
	return []atc.Step{}
}

func createTagList(task manifest.DockerPush, updatePipeline bool) []atc.Step {
	gitRefFile := path.Join(gitDir, ".git", "short_ref")
	versionFile := path.Join(versionName, "version")

	createTagList := &atc.TaskStep{
		Name: "create-tag-list",
		Config: &atc.TaskConfig{
			Platform: "linux",
			ImageResource: &atc.ImageResource{
				Type: "docker-image",
				Source: atc.Source{
					"repository": "alpine",
				},
			},
			Run: atc.TaskRunConfig{
				Path: "/bin/sh",
				Args: []string{"-c", strings.Join([]string{
					fmt.Sprintf("GIT_REF=`[ -f %s ] && cat %s || true`", gitRefFile, gitRefFile),
					fmt.Sprintf("VERSION=`[ -f %s ] && cat %s || true`", versionFile, versionFile),
					fmt.Sprintf("%s > %s", `printf "%s %s" "$GIT_REF" "$VERSION"`, tagListFile),
				}, "\n")},
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitDir},
			},
			Outputs: []atc.TaskOutputConfig{
				{Name: tagList_Dir},
			},
		},
	}
	if updatePipeline {
		createTagList.Config.Inputs = append(createTagList.Config.Inputs, atc.TaskInputConfig{Name: versionName})
	}
	return append([]atc.Step{}, stepWithAttemptsAndTimeout(createTagList, task.GetAttempts(), task.Timeout))
}

func buildAndPushOci(task manifest.DockerPush, resourceName string, fullBasePath string, restore bool, updatePipeline bool) []atc.Step {
	var steps []atc.Step

	params := atc.TaskEnv{
		"CONTEXT":    path.Join(fullBasePath, task.BuildPath),
		"DOCKERFILE": path.Join(fullBasePath, task.DockerfilePath),
	}

	for k, v := range convertVars(task.Vars) {
		params[fmt.Sprintf("BUILD_ARG_%s", k)] = fmt.Sprintf("%s", v)
	}

	buildStep := &atc.TaskStep{
		Name:       "build",
		Privileged: true,
		Config: &atc.TaskConfig{
			Platform: "linux",
			ImageResource: &atc.ImageResource{
				Type: "registry-image",
				Source: atc.Source{
					"repository": "concourse/oci-build-task",
				},
			},
			Params: params,
			Run: atc.TaskRunConfig{
				Path: "build",
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitDir},
			},
			Outputs: []atc.TaskOutputConfig{
				{Name: "image"},
			},
		},
	}
	if restore {
		buildStep.Config.Inputs = append(buildStep.Config.Inputs, atc.TaskInputConfig{Name: dockerBuildTmpDir})
	}

	putStep := &atc.PutStep{
		Name: resourceName,
		Params: atc.Params{
			"image":           "image/image.tar",
			"additional_tags": tagListFile,
		},
	}
	steps = append(steps, stepWithAttemptsAndTimeout(buildStep, task.GetAttempts(), task.GetTimeout()))
	steps = append(steps, createTagList(task, updatePipeline)...)
	steps = append(steps, stepWithAttemptsAndTimeout(putStep, task.GetAttempts(), task.GetTimeout()))
	return steps
}

func buildAndPush(task manifest.DockerPush, resourceName string, ociBuild bool, fullBasePath string, restore bool, updatePipeline bool) []atc.Step {
	if ociBuild {
		return buildAndPushOci(task, resourceName, fullBasePath, restore, updatePipeline)
	}

	step := &atc.PutStep{
		Name: resourceName,
		Params: atc.Params{
			"build":         path.Join(fullBasePath, task.BuildPath),
			"dockerfile":    path.Join(fullBasePath, task.DockerfilePath),
			"tag_as_latest": true,
			"tag_file":      task.GetTagPath(fullBasePath),
			"build_args":    convertVars(task.Vars),
		},
	}
	return append([]atc.Step{}, stepWithAttemptsAndTimeout(step, task.GetAttempts(), task.GetTimeout()))
}

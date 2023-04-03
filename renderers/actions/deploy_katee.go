package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func (a *Actions) deployKateeSteps(task manifest.DeployKatee) (steps Steps) {
	deployKatee := a.createKateeDeployStep(task)
	deploymentStatus := a.createDeploymentStatus(task)
	return append(steps, deployKatee, deploymentStatus)
}

func (a *Actions) createKateeDeployStep(task manifest.DeployKatee) Step {
	deployKatee := Step{
		Name: "Deploy to Katee",
		Uses: "docker://eu.gcr.io/halfpipe-io/ee-katee-vela-cli:latest",
		With: With{
			"entrypoint": "/bin/sh",
			"args":       fmt.Sprintf(`-c "cd %s; /exe vela up -f $KATEE_APPFILE --publish-version $DOCKER_TAG`, a.workingDir)},
		Env: Env{
			"KATEE_TEAM":             strings.TrimPrefix(task.Namespace, "katee-"),
			"KATEE_APPFILE":          task.VelaManifest,
			"KATEE_APPLICATION_NAME": task.ApplicationName,
			"BUILD_VERSION":          "${{ env.BUILD_VERSION }}",
			"GIT_REVISION":           "${{ env.GIT_REVISION }}",
			"KATEE_GKE_CREDENTIALS":  fmt.Sprintf("((%s-service-account-prod.key))", task.Namespace),
		},
	}

	if task.Tag == "gitref" {
		deployKatee.Env["DOCKER_TAG"] = "${{ env.GIT_REVISION }}"
		deployKatee.Env["KATEE_APPLICATION_IMAGE"] = fmt.Sprintf("%s:%s", task.Image, "${{ env.GIT_REVISION }}")
	} else if task.Tag == "version" {
		deployKatee.Env["DOCKER_TAG"] = "${{ env.BUILD_VERSION }}"
		deployKatee.Env["KATEE_APPLICATION_IMAGE"] = fmt.Sprintf("%s:%s", task.Image, "${{ env.BUILD_VERSION }}")
	}

	for k, v := range task.Vars {
		deployKatee.Env[k] = v
	}

	return deployKatee
}

func (a Actions) createDeploymentStatus(task manifest.DeployKatee) Step {
	createDeploymentStatus := Step{
		Name: "Check Deployment Status",
		Uses: "docker://eu.gcr.io/halfpipe-io/ee-katee-vela-cli:latest",
		With: With{
			"entrypoint": "/bin/sh",
			"args":       fmt.Sprintf(`-c "cd %s; /exe deployment-status %s %s $PUBLISHED_VERSION`, a.workingDir, task.Namespace, task.ApplicationName)},
		Env: Env{
			"KATEE_GKE_CREDENTIALS": fmt.Sprintf("((%s-service-account-prod.key))", task.Namespace),
			"KATEE_TEAM":            strings.TrimPrefix(task.Namespace, "katee-"),
		},
	}
	if task.Tag == "gitref" {
		createDeploymentStatus.Env["PUBLISHED_VERSION"] = "${{ env.GIT_REVISION }}"
	} else if task.Tag == "version" {
		createDeploymentStatus.Env["PUBLISHED_VERSION"] = "${{ env.BUILD_VERSION }}"
	}

	return createDeploymentStatus
}

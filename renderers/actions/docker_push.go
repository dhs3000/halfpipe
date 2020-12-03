package actions

import (
	"github.com/springernature/halfpipe/manifest"
	"path"
)

func (a Actions) dockerPushJob(task manifest.DockerPush, man manifest.Manifest) Job {
	basePath := man.Triggers.GetGitTrigger().BasePath
	return Job{
		Name:           task.GetName(),
		RunsOn:         defaultRunner,
		TimeoutMinutes: timeoutInMinutes(task.GetTimeout()),
		Steps: []Step{
			checkoutCode,
			{
				Name: "Set up Docker Buildx",
				Uses: "docker/setup-buildx-action@v1",
			},
			{
				Name: "Login to registry",
				Uses: "docker/login-action@v1",
				With: With{
					{Key: "registry", Value: "eu.gcr.io"},
					{Key: "username", Value: task.Username},
					{Key: "password", Value: task.Password},
				},
			},
			{
				Name: "Build and push",
				Uses: "docker/build-push-action@v2",
				With: With{
					{Key: "context", Value: path.Join(basePath, task.BuildPath)},
					{Key: "file", Value: path.Join(basePath, task.DockerfilePath)},
					{Key: "push", Value: true},
					{Key: "tags", Value: task.Image},
					{Key: "outputs", Value: "type=image,oci-mediatypes=true,push=true"},
				},
			},
			repositoryDispatch(man.PipelineName()),
		},
	}
}

func repositoryDispatch(name string) Step {
	return Step{
		Name: "Repository dispatch",
		Uses: "peter-evans/repository-dispatch@v1",
		With: With{
			{Key: "token", Value: repoAccessToken},
			{Key: "event-type", Value: "docker-push:" + name},
		},
	}

}

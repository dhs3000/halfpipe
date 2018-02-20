package pipeline

import (
	"fmt"
	"path"
	"strings"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/model"
)

type Renderer interface {
	Render(manifest model.Manifest) atc.Config
}

type Pipeline struct{}

func (p Pipeline) gitResource(repo model.Repo) atc.ResourceConfig {
	sources := atc.Source{
		"uri": repo.Uri,
	}

	if repo.PrivateKey != "" {
		sources["private_key"] = repo.PrivateKey
	}

	if len(repo.WatchedPaths) > 0 {
		sources["paths"] = repo.WatchedPaths
	}

	if len(repo.IgnoredPaths) > 0 {
		sources["ignore_paths"] = repo.IgnoredPaths
	}

	if repo.GitCryptKey != "" {
		sources["git_crypt_key"] = repo.GitCryptKey
	}

	return atc.ResourceConfig{
		Name:   repo.GetName(),
		Type:   "git",
		Source: sources,
	}
}

func (p Pipeline) deployCFResource(deployCF model.DeployCF, resourceName string) atc.ResourceConfig {
	sources := atc.Source{
		"api":          deployCF.Api,
		"organization": deployCF.Org,
		"space":        deployCF.Space,
		"username":     deployCF.Username,
		"password":     deployCF.Password,
	}

	return atc.ResourceConfig{
		Name:   resourceName,
		Type:   "cf",
		Source: sources,
	}
}

func (p Pipeline) dockerPushResource(docker model.DockerPush, resourceName string) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name: resourceName,
		Type: "docker-image",
		Source: atc.Source{
			"username":   docker.Username,
			"password":   docker.Password,
			"repository": docker.Repo,
		},
	}
}

func (p Pipeline) imageResource(image string) *atc.ImageResource {
	repo, tag := image, "latest"
	if strings.Contains(image, ":") {
		split := strings.Split(image, ":")
		repo = split[0]
		tag = split[1]
	}
	return &atc.ImageResource{
		Type: "docker-image",
		Source: atc.Source{
			"repository": repo,
			"tag":        tag,
		},
	}
}

func (p Pipeline) runJob(task model.Run, repoName, jobName string) atc.JobConfig {
	return atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
			atc.PlanConfig{
				Task: task.Script,
				TaskConfig: &atc.TaskConfig{
					Platform:      "linux",
					Params:        task.Vars,
					ImageResource: p.imageResource(task.Image),
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Dir:  repoName,
						Args: []string{"-exc", fmt.Sprintf("./%s", strings.Replace(task.Script, "./", "", 1))},
					},
					Inputs: []atc.TaskInputConfig{
						{Name: repoName},
					},
				}}}}
}

func (p Pipeline) deployCFJob(task model.DeployCF, repoName, jobName, resourceName string) atc.JobConfig {
	return atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
			atc.PlanConfig{
				Put: resourceName,
				Params: atc.Params{
					"manifest":              path.Join(repoName, task.Manifest),
					"environment_variables": convertVars(task.Vars),
					"path":                  repoName,
				},
			},
		},
	}
}

func (p Pipeline) dockerPushJob(task model.DockerPush, repoName, jobName, resourceName string) atc.JobConfig {
	return atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
			atc.PlanConfig{Put: resourceName, Params: atc.Params{"build": repoName}},
		},
	}
}

func (p Pipeline) Render(manifest model.Manifest) (config atc.Config) {
	config.Resources = append(config.Resources, p.gitResource(manifest.Repo))
	repoName := manifest.Repo.GetName()

	uniqueName := func(name string) string {
		return getUniqueName(name, &config, 0)
	}

	for i, t := range manifest.Tasks {
		var jobConfig atc.JobConfig
		switch task := t.(type) {
		case model.Run:
			jobName := uniqueName(fmt.Sprintf("run %s", strings.Replace(task.Script, "./", "", 1)))
			jobConfig = p.runJob(task, repoName, jobName)
		case model.DeployCF:
			resourceName := uniqueName(deployCFResourceName(task))
			jobName := uniqueName("deploy-cf")
			config.Resources = append(config.Resources, p.deployCFResource(task, resourceName))
			jobConfig = p.deployCFJob(task, repoName, jobName, resourceName)
		case model.DockerPush:
			resourceName := uniqueName("Docker Registry")
			jobName := uniqueName("docker-push")
			config.Resources = append(config.Resources, p.dockerPushResource(task, resourceName))
			jobConfig = p.dockerPushJob(task, repoName, jobName, resourceName)
		}

		if i > 0 {
			// Plan[0] of a job is ALWAYS the git get.
			jobConfig.Plan[0].Passed = append(jobConfig.Plan[0].Passed, config.Jobs[i-1].Name)
		}
		config.Jobs = append(config.Jobs, jobConfig)
	}
	return
}

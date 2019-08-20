package pipeline

import (
	"strings"

	"regexp"

	"path"

	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

const longResourceCheckInterval = "24h"

func (p pipeline) gitResource(trigger manifest.GitTrigger) atc.ResourceConfig {
	sources := atc.Source{
		"uri": trigger.URI,
	}

	if trigger.PrivateKey != "" {
		sources["private_key"] = trigger.PrivateKey
	}

	if len(trigger.WatchedPaths) > 0 {
		sources["paths"] = trigger.WatchedPaths
	}

	if len(trigger.IgnoredPaths) > 0 {
		sources["ignore_paths"] = trigger.IgnoredPaths
	}

	if trigger.GitCryptKey != "" {
		sources["git_crypt_key"] = trigger.GitCryptKey
	}

	if trigger.Branch == "" {
		sources["branch"] = "master"
	} else {
		sources["branch"] = trigger.Branch
	}

	return atc.ResourceConfig{
		Name:   trigger.GetTriggerName(),
		Type:   "git",
		Source: sources,
	}
}

const slackResourceName = "slack"

func (p pipeline) slackResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name:       slackResourceName,
		Type:       "registry-image",
		CheckEvery: longResourceCheckInterval,
		Source: atc.Source{
			"repository": "cfcommunity/slack-notification-resource",
			"tag":        "v1.4.2",
		},
	}
}

func (p pipeline) slackResource() atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:       slackResourceName,
		Type:       slackResourceName,
		CheckEvery: longResourceCheckInterval,
		Source: atc.Source{
			"url": config.SlackWebhook,
		},
	}
}

func (p pipeline) gcpResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: artifactsResourceName,
		Type: "registry-image",
		Source: atc.Source{
			"repository": config.DockerRegistry + "gcp-resource",
			"tag":        "stable",
			"password":   "((halfpipe-gcr.private_key))",
			"username":   "_json_key",
		},
	}
}

func (p pipeline) artifactResource(man manifest.Manifest) atc.ResourceConfig {
	filter := func(str string) string {
		reg := regexp.MustCompile(`[^a-z0-9\-]+`)
		return reg.ReplaceAllString(strings.ToLower(str), "")
	}

	bucket := config.ArtifactsBucket
	jsonKey := config.ArtifactsJSONKey

	if man.ArtifactConfig.Bucket != "" {
		bucket = man.ArtifactConfig.Bucket
	}
	if man.ArtifactConfig.JSONKey != "" {
		jsonKey = man.ArtifactConfig.JSONKey
	}

	return atc.ResourceConfig{
		Name:       artifactsName,
		Type:       artifactsResourceName,
		CheckEvery: longResourceCheckInterval,
		Source: atc.Source{
			"bucket":   bucket,
			"folder":   path.Join(filter(man.Team), filter(man.PipelineName())),
			"json_key": jsonKey,
		},
	}
}

func (p pipeline) artifactResourceOnFailure(man manifest.Manifest) atc.ResourceConfig {
	config := p.artifactResource(man)
	config.Name = artifactsOnFailureName
	return config
}

func (p pipeline) cronResource(trigger manifest.CronTrigger) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:       cronName,
		Type:       "cron-resource",
		CheckEvery: "1m",
		Source: atc.Source{
			"expression":       trigger.Trigger,
			"location":         "UTC",
			"fire_immediately": true,
		},
	}
}

func cronResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name:                 "cron-resource",
		Type:                 "registry-image",
		UniqueVersionHistory: true,
		Source: atc.Source{
			"repository": "cftoolsmiths/cron-resource",
			"tag":        "v0.3",
		},
	}
}

func halfpipeCfDeployResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: "cf-resource",
		Type: "registry-image",
		Source: atc.Source{
			"repository": config.DockerRegistry + "cf-resource",
			"tag":        "stable",
			"password":   "((halfpipe-gcr.private_key))",
			"username":   "_json_key",
		},
	}
}

func (p pipeline) deployCFResource(deployCF manifest.DeployCF, resourceName string) atc.ResourceConfig {
	sources := atc.Source{
		"api":      deployCF.API,
		"org":      deployCF.Org,
		"space":    deployCF.Space,
		"username": deployCF.Username,
		"password": deployCF.Password,
	}

	return atc.ResourceConfig{
		Name:       resourceName,
		Type:       "cf-resource",
		Source:     sources,
		CheckEvery: longResourceCheckInterval,
	}
}

func (p pipeline) dockerPushResource(docker manifest.DockerPush) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name: dockerPushResourceName(docker),
		Type: "docker-image",
		Source: atc.Source{
			"username":   docker.Username,
			"password":   docker.Password,
			"repository": docker.Image,
		},
		CheckEvery: longResourceCheckInterval,
	}
}

func (p pipeline) imageResource(docker manifest.Docker) *atc.ImageResource {
	repo, tag := docker.Image, "latest"
	if strings.Contains(docker.Image, ":") {
		split := strings.Split(docker.Image, ":")
		repo = split[0]
		tag = split[1]
	}

	source := atc.Source{
		"repository": repo,
		"tag":        tag,
	}

	if docker.Username != "" && docker.Password != "" {
		source["username"] = docker.Username
		source["password"] = docker.Password
	}

	return &atc.ImageResource{
		Type:   "registry-image",
		Source: source,
	}
}

func (p pipeline) versionResource(manifest manifest.Manifest) atc.ResourceConfig {
	key := fmt.Sprintf("%s-%s", manifest.Team, manifest.Pipeline)
	gitTrigger := manifest.Triggers.GetGitTrigger()
	if gitTrigger.Branch != "" && gitTrigger.Branch != "master" {
		key = fmt.Sprintf("%s-%s", key, gitTrigger.Branch)
	}

	return atc.ResourceConfig{
		Name: versionName,
		Type: "semver",
		Source: atc.Source{
			"driver":   "gcs",
			"key":      key,
			"bucket":   config.VersionBucket,
			"json_key": config.VersionJSONKey,
		},
	}
}

func (p pipeline) updateJobConfig(manifest manifest.Manifest, basePath string) *atc.JobConfig {
	return &atc.JobConfig{
		Name:   updateJobName,
		Serial: true,
		Plan: []atc.PlanConfig{
			p.updatePipelineTask(manifest, basePath),
			{
				Put: versionName,
				Params: atc.Params{
					"bump": "minor",
				},
				Attempts: updateTaskAttempts,
			}},
	}
}

func (p pipeline) updatePipelineTask(man manifest.Manifest, basePath string) atc.PlanConfig {
	return atc.PlanConfig{
		Task:     updatePipelineName,
		Attempts: updateTaskAttempts,
		TaskConfig: &atc.TaskConfig{
			Platform: "linux",
			Params: map[string]string{
				"CONCOURSE_URL":      "((concourse.url))",
				"CONCOURSE_PASSWORD": "((concourse.password))",
				"CONCOURSE_TEAM":     "((concourse.team))",
				"CONCOURSE_USERNAME": "((concourse.username))",
				"PIPELINE_NAME":      man.PipelineName(),
				"HALFPIPE_DOMAIN":    config.Domain,
				"HALFPIPE_PROJECT":   config.Project,
			},
			ImageResource: p.imageResource(manifest.Docker{
				Image:    config.DockerRegistry + "halfpipe-auto-update",
				Username: "_json_key",
				Password: "((halfpipe-gcr.private_key))",
			}),
			Run: atc.TaskRunConfig{
				Path: "update-pipeline",
				Dir:  path.Join(gitDir, basePath),
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitName},
			},
		}}
}

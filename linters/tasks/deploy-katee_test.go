package tasks

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKateeDeployTaskWithEmptyTask(t *testing.T) {
	task := manifest.DeployKatee{VelaAppFile: "vela.yml"}
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors, warnings := LintDeployKateeTask(task, emptyManifest, fs)
	assert.Len(t, errors, 2)
	assert.Len(t, warnings, 0)

	linterrors.AssertMissingFieldInErrors(t, "applicationName", errors)
	linterrors.AssertFileErrorInErrors(t, "vela.yml", errors)
}

func TestKateeDeployRetries(t *testing.T) {
	task := manifest.DeployKatee{}

	task.Retries = -1
	errors, _ := LintDeployKateeTask(task, emptyManifest, afero.Afero{Fs: afero.NewMemMapFs()})
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 6
	errors, _ = LintDeployKateeTask(task, emptyManifest, afero.Afero{Fs: afero.NewMemMapFs()})
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)
}

func TestKateeDeployTag(t *testing.T) {
	task := manifest.DeployKatee{ApplicationName: "app", VelaAppFile: "vela.yml"}
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("vela.yml", []byte("foo"), 0777)

	t.Run("not set", func(t *testing.T) {
		errors, warnings := LintDeployKateeTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	})

	t.Run("gitref", func(t *testing.T) {
		task.Tag = "gitref"
		errors, warnings := LintDeployKateeTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	})

	t.Run("version with update-pipeline feature", func(t *testing.T) {
		manifestWithUpdate := manifest.Manifest{FeatureToggles: manifest.FeatureToggles{manifest.FeatureUpdatePipeline}}
		task.Tag = "version"
		errors, warnings := LintDeployKateeTask(task, manifestWithUpdate, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	})

	t.Run("version without update-pipeline feature", func(t *testing.T) {
		manifestConcourse := manifest.Manifest{Platform: "concourse"}
		task.Tag = "version"
		errors, warnings := LintDeployKateeTask(task, manifestConcourse, fs)
		assert.Len(t, errors, 1)
		assert.Len(t, warnings, 0)
	})

	t.Run("version without update-pipeline feature", func(t *testing.T) {
		manifestActions := manifest.Manifest{Platform: "actions"}
		task.Tag = "version"
		errors, warnings := LintDeployKateeTask(task, manifestActions, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	})

	t.Run("invalid", func(t *testing.T) {
		task.Tag = "bananas"
		errors, warnings := LintDeployKateeTask(task, emptyManifest, fs)
		assert.Len(t, errors, 1)
		linterrors.AssertInvalidFieldInErrors(t, "tag", errors)
		assert.Len(t, warnings, 0)
	})
}

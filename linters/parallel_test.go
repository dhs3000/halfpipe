package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestParallelTaskInParallelTask(t *testing.T) {
	task := manifest.Parallel{
		Tasks: manifest.TaskList{
			manifest.Run{},
			manifest.DockerPush{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{},
				},
			},
		},
	}
	errs, _ := LintParallelTask(task)
	AssertContainsError(t, errs, ErrInvalidField.WithValue("type"))
}

func TestErrorIfParallelIsEmpty(t *testing.T) {
	task := manifest.Parallel{
		Tasks: manifest.TaskList{},
	}
	errs, warns := LintParallelTask(task)
	assert.Len(t, errs, 1)
	assert.Len(t, warns, 0)
	AssertContainsError(t, errs, ErrInvalidField.WithValue("tasks"))
}

func TestWarningIfParallelOnlyContainsOneItem(t *testing.T) {
	task := manifest.Parallel{
		Tasks: manifest.TaskList{
			manifest.Run{},
		},
	}
	errs, warns := LintParallelTask(task)
	assert.Len(t, errs, 0)
	assert.Len(t, warns, 1)
	AssertContainsError(t, warns, ErrInvalidField.WithValue("tasks"))
}

func TestWarnIfMultipleTasksInsideParallelSavesArtifact(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		task := manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{SaveArtifactsOnFailure: []string{"blah"}},
				manifest.Run{SaveArtifacts: []string{"."}},
				manifest.Run{},
			},
		}

		errs, warns := LintParallelTask(task)
		assert.Len(t, errs, 0)
		assert.Len(t, warns, 0)
	})

	t.Run("bad", func(t *testing.T) {
		task := manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Run{SaveArtifacts: []string{"."}},
				manifest.Run{},
				manifest.Run{},
				manifest.Run{SaveArtifacts: []string{"."}},
			},
		}

		errs, warns := LintParallelTask(task)
		assert.Len(t, errs, 0)
		assert.Len(t, warns, 1)
		AssertContainsError(t, warns, ErrInvalidField.WithValue("tasks"))
	})
}

func TestWarnIfMultipleTasksInsideParallelSavesArtifactOnFailure(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		task := manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Run{SaveArtifacts: []string{"."}},
				manifest.Run{},
			},
		}

		errs, warns := LintParallelTask(task)
		assert.Len(t, errs, 0)
		assert.Len(t, warns, 0)
	})

	t.Run("bad", func(t *testing.T) {
		task := manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Run{SaveArtifacts: []string{"."}, SaveArtifactsOnFailure: []string{"blurgh"}},
				manifest.Run{},
				manifest.Run{},
				manifest.Run{SaveArtifactsOnFailure: []string{"."}},
			},
		}

		errs, warns := LintParallelTask(task)
		assert.Len(t, errs, 0)
		assert.Len(t, warns, 1)
		AssertContainsError(t, warns, ErrInvalidField.WithValue("tasks"))
	})
}

package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestSeqMustComeFromAParallelTask(t *testing.T) {
	errs, warnings := LintSequenceTask(manifest.Sequence{Tasks: manifest.TaskList{manifest.Run{}, manifest.Run{}}}, false)

	AssertContainsError(t, errs, ErrInvalidField.WithValue("type"))
	assert.Empty(t, warnings)
}

func TestSeqIsAtLeastOne(t *testing.T) {
	t.Run("errors with empty sequence", func(t *testing.T) {
		errs, warnings := LintSequenceTask(manifest.Sequence{}, true)

		AssertContainsError(t, errs, ErrInvalidField.WithValue("tasks"))
		assert.Empty(t, warnings)
	})

	t.Run("ok with two task", func(t *testing.T) {
		errs, warnings := LintSequenceTask(manifest.Sequence{Tasks: manifest.TaskList{manifest.Run{}, manifest.Run{}}}, true)

		assert.Empty(t, errs)
		assert.Empty(t, warnings)
	})
}

func TestSeqDoesNotContainOtherSeqsOrParallels(t *testing.T) {
	t.Run("errors when sequence contains sequence", func(t *testing.T) {
		errs, warnings := LintSequenceTask(manifest.Sequence{
			Type: "",
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Sequence{},
			},
		}, true)
		assert.Len(t, errs, 1)
		assert.Len(t, warnings, 0)
		AssertContainsError(t, errs, ErrInvalidField.WithValue("tasks"))
	})
}

package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestTeamIsDifferentFromTriggerTeam(t *testing.T) {
	trigger := manifest.PipelineTrigger{
		Team: "team-a",
	}

	man := manifest.Manifest{
		Team: "team-b",
		Triggers: manifest.TriggerList{
			trigger,
		},
	}
	errs, warns := LintPipelineTrigger(man, trigger)

	assert.Len(t, errs, 1)
	assert.Len(t, warns, 0)
	assertContainsError(t, errs, ErrInvalidField.WithValue("team"))
}

func TestEmptyPipeline(t *testing.T) {
	trigger := manifest.PipelineTrigger{
		Team: "team",
	}

	man := manifest.Manifest{
		Team: "team",
		Triggers: manifest.TriggerList{
			trigger,
		},
	}

	errs, warns := LintPipelineTrigger(man, trigger)

	assert.Len(t, errs, 1)
	assert.Len(t, warns, 0)
	assertContainsError(t, errs, ErrInvalidField.WithValue("pipeline"))
}

func TestEmptyJob(t *testing.T) {
	trigger := manifest.PipelineTrigger{
		Team:     "team",
		Pipeline: "asd",
	}

	man := manifest.Manifest{
		Team: "team",
		Triggers: manifest.TriggerList{
			trigger,
		},
	}

	errs, warns := LintPipelineTrigger(man, trigger)

	assert.Len(t, errs, 1)
	assert.Len(t, warns, 0)
	assertContainsError(t, errs, ErrInvalidField.WithValue("job"))
}

func TestBadStatus(t *testing.T) {
	trigger := manifest.PipelineTrigger{
		Team:     "team",
		Pipeline: "asd",
		Job:      "asdf",
		Status:   "kehe",
	}

	man := manifest.Manifest{
		Team: "team",
		Triggers: manifest.TriggerList{
			trigger,
		},
	}

	errs, warns := LintPipelineTrigger(man, trigger)

	assert.Len(t, errs, 1)
	assert.Len(t, warns, 0)
	assertContainsError(t, errs, ErrInvalidField.WithValue("status"))
}

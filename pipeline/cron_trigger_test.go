package pipeline

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestCronTriggerResourceTypeSet(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.TimerTrigger{
				Cron: "*/10 * * * *",
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "run.sh"},
		},
	}

	config := testPipeline().Render(man)
	_, found := config.ResourceTypes.Lookup("cron-resource")
	assert.True(t, found)
}

func TestCronTriggerNotSet(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},
		Tasks: []manifest.Task{
			manifest.Run{Name: "blah", Script: "run.sh"},
		},
	}
	config := testPipeline().Render(man)
	resources := config.Resources
	plan := config.Jobs[0].Plan

	//should be 1 resource: git
	assert.Len(t, resources, 1)
	assert.Equal(t, "git", resources[0].Type)

	//should be 2 items in the plan: get git + task
	assert.Len(t, plan, 2)
	assert.Equal(t, manifest.GitTrigger{}.GetTriggerName(), (plan[0].InParallel.Steps)[0].Name())
	assert.True(t, (plan[0].InParallel.Steps)[0].Trigger)
	assert.Equal(t, "blah", plan[1].Task)
}

func TestCronTriggerSetAddsResource(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.TimerTrigger{
				Cron: "*/10 * * * *",
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "run.sh"},
		},
	}

	config := testPipeline().Render(man)
	resource, found := config.Resources.Lookup(manifest.TimerTrigger{}.GetTriggerName())
	assert.True(t, found)
	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), resource.Name)
	assert.Equal(t, "cron-resource", resource.Type)
	assert.Equal(t, man.Triggers[0].(manifest.TimerTrigger).Cron, resource.Source["expression"])
	assert.Equal(t, "1m", resource.CheckEvery)
}

func TestCronTriggerSetWithCorrectPassedOnSecondJob(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.TimerTrigger{
				Cron: "*/10 * * * *",
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "s1.sh"},
			manifest.Run{Script: "s2.sh"},
		},
	}
	config := testPipeline().Render(man)

	t1 := config.Jobs[0]
	t1InParallel := t1.Plan[0].InParallel.Steps

	assert.Len(t, t1.Plan, 2)
	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), t1InParallel[0].Name())
	assert.True(t, t1InParallel[0].Trigger)

	t2 := config.Jobs[1]
	t2InParallel := t2.Plan[0].InParallel.Steps
	assert.Len(t, t2.Plan, 2)

	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), t2InParallel[0].Name())
	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerAttempts(), t2InParallel[0].Attempts)
	assert.Equal(t, []string{t1.Name}, t2InParallel[0].Passed)
}

func TestCronTriggerSetWithParallelTasks(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.TimerTrigger{
				Cron: "*/10 * * * *",
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "first.sh"},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Script: "p1.sh"},
					manifest.Run{Script: "p2.sh"},
				},
			},
			manifest.Run{Script: "last.sh"},
		},
	}
	config := testPipeline().Render(man)

	first := config.Jobs[0]
	firstInParallel := first.Plan[0].InParallel.Steps

	assert.Len(t, first.Plan, 2)
	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), firstInParallel[0].Name())
	assert.True(t, firstInParallel[0].Trigger)

	p1 := config.Jobs[1]
	p1InParallel := p1.Plan[0].InParallel.Steps
	assert.Len(t, p1.Plan, 2)

	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), p1InParallel[0].Name())
	assert.Equal(t, []string{first.Name}, p1InParallel[0].Passed)

	p2 := config.Jobs[2]
	p2InParallel := p2.Plan[0].InParallel.Steps
	assert.Len(t, p2.Plan, 2)

	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), p2InParallel[0].Name())
	assert.Equal(t, []string{first.Name}, p2InParallel[0].Passed)

	last := config.Jobs[3].Plan
	lastInParallel := last[0].InParallel.Steps
	assert.Len(t, last, 2)

	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), lastInParallel[0].Name())
	assert.Equal(t, []string{p1.Name, p2.Name}, lastInParallel[0].Passed)
}

func TestCronTriggerSetWhenUsingRestoreArtifact(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.TimerTrigger{
				Cron: "*/10 * * * *",
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "first.sh", SaveArtifacts: []string{"something"}},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Script: "p1.sh"},
					manifest.Run{Script: "p2.sh", RestoreArtifacts: true},
				},
			},
			manifest.Run{Script: "last.sh", RestoreArtifacts: true},
		},
	}

	config := testPipeline().Render(man)

	first := config.Jobs[0]
	firstInParallel := first.Plan[0].InParallel.Steps

	assert.Len(t, first.Plan, 3)
	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), firstInParallel[0].Name())
	assert.True(t, firstInParallel[0].Trigger)

	p1 := config.Jobs[1]
	p1InParallel := p1.Plan[0].InParallel.Steps
	assert.Len(t, p1.Plan, 2)

	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), p1InParallel[0].Name())
	assert.Equal(t, []string{first.Name}, p1InParallel[0].Passed)

	p2 := config.Jobs[2]
	p2InParallel := p2.Plan[0].InParallel.Steps
	assert.Len(t, p2.Plan, 3)

	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), p2InParallel[0].Name())
	assert.Equal(t, []string{first.Name}, p2InParallel[0].Passed)

	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), p2InParallel[0].Name())
	assert.Equal(t, []string{first.Name}, p2InParallel[0].Passed)
	assert.Equal(t, restoreArtifactTask(man), p2.Plan[1])

	last := config.Jobs[3].Plan
	lastInParallel := last[0].InParallel.Steps
	assert.Len(t, last, 3)

	assert.Equal(t, manifest.TimerTrigger{}.GetTriggerName(), lastInParallel[0].Name())
	assert.Equal(t, []string{p1.Name, p2.Name}, lastInParallel[0].Passed)

	// Artifacts should not have any passed.
	assert.Equal(t, restoreArtifactTask(man), last[1])
}

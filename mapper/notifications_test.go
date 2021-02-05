package mapper

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoesNothingWhenSlackChannelIsNotDefined(t *testing.T) {
	updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{})
	assert.Equal(t, manifest.Manifest{}, updated)
}

func TestUpdatesNotificationsWhenSlackChannelIsDefined(t *testing.T) {
	t.Run("Only failure", func(t *testing.T) {
		input := manifest.Manifest{
			SlackChannel: "#test",
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.DockerPush{},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.DeployMLZip{},
						manifest.Sequence{
							Tasks: manifest.TaskList{
								manifest.DeployCF{},
								manifest.ConsumerIntegrationTest{},
							},
						},
					},
				},
			},
		}

		notifications := manifest.Notifications{OnFailure: []string{input.SlackChannel}}
		expected := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{Notifications: notifications},
				manifest.DockerPush{Notifications: notifications},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.DeployMLZip{Notifications: notifications},
						manifest.Sequence{
							Tasks: manifest.TaskList{
								manifest.DeployCF{Notifications: notifications},
								manifest.ConsumerIntegrationTest{Notifications: notifications},
							},
						},
					},
				},
			},
		}

		updated, _ := NewNotificationsMapper().Apply(input)
		assert.Equal(t, expected, updated)

		// Make sure we dont update the old manifest in place, cus that leads to horrible bugs.
		assert.NotEqual(t, updated, input)
	})

	t.Run("Doesn't update the cf push pre-promotes", func(t *testing.T) {
		input := manifest.Manifest{
			SlackChannel: "#test",
			Tasks: manifest.TaskList{
				manifest.DeployCF{},
				manifest.DeployCF{
					PrePromote: manifest.TaskList{
						manifest.Run{},
						manifest.Run{},
					},
				},
				manifest.DeployCF{},
			},
		}

		notifications := manifest.Notifications{OnFailure: []string{input.SlackChannel}}
		expected := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.DeployCF{Notifications: notifications},
				manifest.DeployCF{
					Notifications: notifications,
					PrePromote: manifest.TaskList{
						manifest.Run{},
						manifest.Run{},
					},
				},
				manifest.DeployCF{
					Notifications: notifications,
				},
			},
		}

		updated, _ := NewNotificationsMapper().Apply(input)
		assert.Equal(t, expected, updated)
	})

	t.Run("Both failure and success", func(t *testing.T) {
		input := manifest.Manifest{
			SlackChannel: "#test",
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.DockerPush{NotifyOnSuccess: true},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.DeployMLZip{NotifyOnSuccess: true},
						manifest.Sequence{
							Tasks: manifest.TaskList{
								manifest.DeployCF{NotifyOnSuccess: true},
								manifest.ConsumerIntegrationTest{},
							},
						},
					},
				},
			},
		}

		notifications := manifest.Notifications{OnFailure: []string{input.SlackChannel}}
		notificationsWithSuccess := manifest.Notifications{OnFailure: []string{input.SlackChannel}, OnSuccess: []string{input.SlackChannel}}
		expected := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{Notifications: notifications},
				manifest.DockerPush{NotifyOnSuccess: true, Notifications: notificationsWithSuccess},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.DeployMLZip{NotifyOnSuccess: true, Notifications: notificationsWithSuccess},
						manifest.Sequence{
							Tasks: manifest.TaskList{
								manifest.DeployCF{NotifyOnSuccess: true, Notifications: notificationsWithSuccess},
								manifest.ConsumerIntegrationTest{Notifications: notifications},
							},
						},
					},
				},
			},
		}

		updated, _ := NewNotificationsMapper().Apply(input)
		assert.Equal(t, expected, updated)
	})

	t.Run("Doesnt map if Notifications is already defined", func(t *testing.T) {
		specialSnowflake1 := manifest.Notifications{
			OnSuccess: []string{"1"},
			OnFailure: []string{"2"},
		}

		specialSnowflake2 := manifest.Notifications{
			OnSuccess: []string{"a", "b"},
			OnFailure: []string{"x", "y", "z"},
		}

		input := manifest.Manifest{
			SlackChannel: "#test",
			Tasks: manifest.TaskList{
				manifest.Run{Notifications: specialSnowflake1},
				manifest.DockerPush{},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.DeployMLZip{NotifyOnSuccess: true},
						manifest.Sequence{
							Tasks: manifest.TaskList{
								manifest.DeployCF{NotifyOnSuccess: true},
								manifest.ConsumerIntegrationTest{Notifications: specialSnowflake2},
							},
						},
					},
				},
			},
		}

		notifications := manifest.Notifications{OnFailure: []string{input.SlackChannel}}
		notificationsWithSuccess := manifest.Notifications{OnFailure: []string{input.SlackChannel}, OnSuccess: []string{input.SlackChannel}}
		expected := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{Notifications: specialSnowflake1},
				manifest.DockerPush{Notifications: notifications},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.DeployMLZip{NotifyOnSuccess: true, Notifications: notificationsWithSuccess},
						manifest.Sequence{
							Tasks: manifest.TaskList{
								manifest.DeployCF{NotifyOnSuccess: true, Notifications: notificationsWithSuccess},
								manifest.ConsumerIntegrationTest{Notifications: specialSnowflake2},
							},
						},
					},
				},
			},
		}

		updated, _ := NewNotificationsMapper().Apply(input)
		assert.Equal(t, expected, updated)
	})
}

func TestDefaultNotificationMessages(t *testing.T) {
	defaultFailureMessage := "failure msg"
	defaultSuccessMessage := "success msg"

	input := manifest.Manifest{
		SlackChannel:        "#test",
		SlackFailureMessage: defaultFailureMessage,
		SlackSuccessMessage: defaultSuccessMessage,
		Tasks: manifest.TaskList{
			manifest.Run{},
			manifest.DockerPush{},
			manifest.DeployCF{
				Notifications: manifest.Notifications{
					OnSuccess:        []string{"#foo"},
					OnSuccessMessage: "custom",
					OnFailureMessage: "custom",
				},
			},
		},
	}

	updated, _ := NewNotificationsMapper().Apply(input)
	assert.Equal(t, defaultFailureMessage, updated.Tasks[0].GetNotifications().OnFailureMessage)
	assert.Equal(t, defaultFailureMessage, updated.Tasks[1].GetNotifications().OnFailureMessage)
	assert.Equal(t, "custom", updated.Tasks[2].GetNotifications().OnFailureMessage)

	assert.Equal(t, defaultSuccessMessage, updated.Tasks[0].GetNotifications().OnSuccessMessage)
	assert.Equal(t, defaultSuccessMessage, updated.Tasks[1].GetNotifications().OnSuccessMessage)
	assert.Equal(t, "custom", updated.Tasks[2].GetNotifications().OnFailureMessage)

}

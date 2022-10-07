package linters

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

var WarnScriptMustExistInDockerImage = func(script string) error {
	return fmt.Errorf("make sure '%s' is available in the docker image you have specified", script)
}

var WarnMakeSureScriptIsExecutable = func(script string) error {
	return fmt.Errorf("we have disabled the executable test for windows hosts. Make sure '%s' is actually executable otherwise the pipeline will produce runtime errors", script)
}

func LintRunTask(run manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) {
	if run.Script == "" {
		errs = append(errs, NewErrMissingField("script"))
	} else if strings.HasPrefix(run.Script, `\`) {
		command := strings.Fields(strings.TrimSpace(run.Script))[0]
		commandWithoutSlashPrefix := command[1:]
		warnings = append(warnings, WarnScriptMustExistInDockerImage(commandWithoutSlashPrefix))
	} else {
		// Possible for script to have args,
		fields := strings.Fields(strings.TrimSpace(run.Script))
		command := fields[0]
		checkForExecutable := os != "windows"
		if !checkForExecutable {
			warnings = append(warnings, WarnMakeSureScriptIsExecutable(command))
		}
		if err := CheckFile(fs, command, checkForExecutable); err != nil {
			errs = append(errs, err)
		}
	}

	if run.Retries < 0 || run.Retries > 5 {
		errs = append(errs, NewErrInvalidField("retries", "must be between 0 and 5"))
	}

	if run.Docker.Image == "" {
		errs = append(errs, NewErrMissingField("docker.image"))
	}

	if run.Docker.Username != "" && run.Docker.Password == "" {
		errs = append(errs, NewErrMissingField("docker.password"))
	}
	if run.Docker.Password != "" && run.Docker.Username == "" {
		errs = append(errs, NewErrMissingField("docker.username"))
	}

	return errs, warnings
}

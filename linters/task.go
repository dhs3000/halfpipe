package linters

import (
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
)

type TaskLinter struct{}

func (t TaskLinter) Lint(man model.Manifest) []error {
	var errs []error
	if len(man.Tasks) == 0 {
		errs = append(errs, errors.NewMissingField("tasks"))
		return errs
	}

	return errs
}

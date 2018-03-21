package linters

import "fmt"
import (
	"strings"

	"path"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/linters/errors"
)

type LintResults []LintResult
type LintResult struct {
	Linter   string
	Errors   []error
	Warnings []error
}

func NewLintResult(linter string, errs []error, warns []error) LintResult {
	return LintResult{
		Linter:   linter,
		Errors:   errs,
		Warnings: warns,
	}
}

func (lrs LintResults) HasWarnings() bool {
	for _, lintResult := range lrs {
		if lintResult.HasWarnings() {
			return true
		}
	}
	return false
}

func (lrs LintResults) HasErrors() bool {
	for _, lintResult := range lrs {
		if lintResult.HasErrors() {
			return true
		}
	}
	return false
}

func (lrs LintResults) Error() (out string) {
	for _, result := range lrs {
		out += result.Error()
		out += "\n"
	}
	return
}

func (lr LintResult) Error() (out string) {
	out += fmt.Sprintf("%s\n", lr.Linter)

	if lr.HasErrors() {
		out += formatErrors("Errors", lr.Errors)
	}
	if lr.HasWarnings() {
		out += formatErrors("Warnings", lr.Warnings)
	}

	if !lr.HasErrors() && !lr.HasWarnings() {
		out += fmt.Sprintf("\t%s\n\n", `No issues \o/`)
	}

	return
}

func formatErrors(typeOfError string, errs []error) (out string) {
	out += fmt.Sprintf("\t%s:\n", typeOfError)
	for _, err := range deduplicate(errs) {
		out += fmt.Sprintf("\t\t* %s\n", err)
		if doc, ok := err.(errors.Documented); ok {
			out += fmt.Sprintf("\t  see: %s\n", renderDocLink(doc.DocID()))
		}
	}
	return
}

func (lr LintResult) HasErrors() bool {
	return len(lr.Errors) != 0
}

func (lr LintResult) HasWarnings() bool {
	return len(lr.Warnings) != 0
}

func (lr *LintResult) AddError(err ...error) {
	lr.Errors = append(lr.Errors, err...)
}

func (lr *LintResult) AddWarning(err ...error) {
	lr.Warnings = append(lr.Warnings, err...)
}

func deduplicate(errs []error) (errors []error) {
	for _, err := range errs {
		if !errorInErrors(err, errors) {
			errors = append(errors, err)
		}
	}
	return
}

func errorInErrors(err error, errs []error) bool {
	for _, e := range errs {
		if err == e {
			return true
		}
	}
	return false
}

func renderDocLink(docID string) string {
	return path.Join(config.DocHost, "/docs/cli-errors", renderDocAnchor(docID))
}

func renderDocAnchor(docID string) string {
	if docID != "" {
		return fmt.Sprintf("#%s", normalize(docID))
	}
	return ""
}

func normalize(value string) string {
	withoutWhiteSpace := strings.Replace(strings.ToLower(value), " ", "-", -1)
	return strings.Replace(withoutWhiteSpace, ".", "", -1)

}

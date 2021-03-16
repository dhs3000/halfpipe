package filechecker

import (
	"github.com/springernature/halfpipe/config"
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/stretchr/testify/assert"
)

func testFs() afero.Afero {
	return afero.Afero{Fs: afero.NewMemMapFs()}
}

func TestFile_NotExists(t *testing.T) {
	fs := testFs()
	err := CheckFile(fs, "missing.file", false)

	assert.Equal(t, linterrors.NewFileError("missing.file", "does not exist"), err)
}

func TestFile_Empty(t *testing.T) {
	fs := testFs()
	fs.WriteFile(".halfpipe.io", []byte{}, 0777)

	err := CheckFile(fs, ".halfpipe.io", false)
	assert.Equal(t, linterrors.NewFileError(".halfpipe.io", "is empty"), err)
}

func TestFile_IsDirectory(t *testing.T) {
	fs := testFs()
	fs.Mkdir("build", 0777)

	err := CheckFile(fs, "build", false)
	assert.Equal(t, linterrors.NewFileError("build", "is not a file"), err)
}

func TestFile_NotExecutable(t *testing.T) {
	fs := testFs()
	fs.WriteFile("build.sh", []byte("go test"), 0400)

	err := CheckFile(fs, "build.sh", true)
	assert.Equal(t, linterrors.NewFileError("build.sh", "is not executable"), err)

	err = CheckFile(fs, "build.sh", false)
	assert.Nil(t, err)
}

func TestFile_Happy(t *testing.T) {
	fs := testFs()
	fs.WriteFile(".halfpipe.io", []byte("foo"), 0700)
	err := CheckFile(fs, ".halfpipe.io", true)

	assert.Nil(t, err)
}

func TestRead(t *testing.T) {
	fs := testFs()
	fs.WriteFile(".halfpipe.io", []byte("foo"), 0700)

	content, err := ReadFile(fs, ".halfpipe.io")

	assert.Nil(t, err)
	assert.Equal(t, "foo", content)
}

func TestReadDoesCheck(t *testing.T) {
	fs := testFs()
	fs.WriteFile(".halfpipe.io", []byte{}, 0700)

	_, err := ReadFile(fs, ".halfpipe.io")

	assert.Equal(t, linterrors.NewFileError(".halfpipe.io", "is empty"), err)
}

func TestReadHalfpipeFilesErrorsTwoFileOptionsExist(t *testing.T) {
	fs := testFs()
	fs.WriteFile(".halfpipe.io", []byte{}, 0700)
	fs.WriteFile(".halfpipe.io.yml", []byte{}, 0700)

	_, err := GetHalfpipeFileName(fs, "", config.HalfpipeFilenameOptions)

	assert.EqualError(t, err, "found [.halfpipe.io .halfpipe.io.yml] files. Please use only 1 of those")
}

func TestReadHalfpipeFilesErrorsWhenBothOptionsAreNotThere(t *testing.T) {
	pr := testFs()

	_, err := GetHalfpipeFileName(pr, "", config.HalfpipeFilenameOptions)

	assert.EqualError(t, err, "couldn't find any of the allowed [.halfpipe.io .halfpipe.io.yml .halfpipe.io.yaml] files")
}

func TestReadHalfpipeFilesErrorsWhenExplicitFilenameGivenButFIleIsMissing(t *testing.T) {
	pr := testFs()

	_, err := GetHalfpipeFileName(pr, "", []string{"some-other-file.yml"})

	assert.EqualError(t, err, "couldn't find 'some-other-file.yml'")
}

func TestReadHalfpipeFilesIsHappyWithOneOfTheOptions(t *testing.T) {
	fs := testFs()
	fs.WriteFile(".halfpipe.io", []byte("foo"), 0700)
	_, err := GetHalfpipeFileName(fs, "", config.HalfpipeFilenameOptions)
	assert.Nil(t, err)
}

func TestReadHalfpipeFilesIsHappyWithExplicitFilenameGiven(t *testing.T) {
	fs := testFs()
	fs.WriteFile("some-other-file.yml", []byte("foo"), 0700)
	_, err := GetHalfpipeFileName(fs, "", []string{"some-other-file.yml"})
	assert.Nil(t, err)
}

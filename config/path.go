package config

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Path represents an absolute filesystem path.
type Path string

// FilePath is a special-case of Path representing a file.
type FilePath Path

// DirectoryPath is a special-case of Path representing a directory.
type DirectoryPath Path

// UnmarshalText parses a generic path. It checks that the base
// directory for this path exists.
func (p *Path) UnmarshalText(text []byte) error {
	absolute, err := filepath.Abs(string(text))
	if err != nil {
		return errors.Wrapf(err, "cannot get absolute path for %q", string(text))
	}
	_, err = os.Stat(filepath.Dir(absolute))
	if err != nil {
		return errors.Wrapf(err, "%q is in a non-existant directory", absolute)
	}
	*p = Path(absolute)
	return nil
}

// UnmarshalText parses a file path.
func (p *FilePath) UnmarshalText(text []byte) error {
	path := Path("")
	err := path.UnmarshalText(text)
	if err != nil {
		return err
	}
	stat, err := os.Stat(string(path))
	if err != nil {
		*p = FilePath(path)
		return nil
	}
	if stat.IsDir() {
		return errors.Errorf("%q is a directory", string(path))
	}
	*p = FilePath(path)
	return nil
}

// UnmarshalText parses a directory path.
func (p *DirectoryPath) UnmarshalText(text []byte) error {
	path := Path("")
	err := path.UnmarshalText(text)
	if err != nil {
		return err
	}
	stat, err := os.Stat(string(path))
	if err != nil {
		*p = DirectoryPath(path)
		return nil
	}
	if !stat.IsDir() {
		return errors.Errorf("%q is a regular file", string(path))
	}
	*p = DirectoryPath(path)
	return nil
}

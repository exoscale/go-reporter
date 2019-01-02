package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func inPreparedDirectory(t *testing.T, f func(dir string)) {
	tempDir, err := ioutil.TempDir("", "paths")
	if err != nil {
		t.Fatalf("Cannot create temporary directory:\n%+v", err)
	}
	defer os.RemoveAll(tempDir)
	if err := os.Mkdir(path.Join(tempDir, "batman"), 0700); err != nil {
		t.Fatalf("Cannot create directory %q:\n%+v", "batman", err)
	}
	if err := ioutil.WriteFile(path.Join(tempDir, "robin"), []byte{}, 0600); err != nil {
		t.Fatalf("Cannot create file %q:\n%+v", "robin", err)
	}

	f(tempDir)
}

func TestUnmarshalFilePath(t *testing.T) {
	inPreparedDirectory(t, func(dir string) {
		cases := []struct {
			in   string
			want string
		}{
			{
				// Existing file
				in:   fmt.Sprintf("%s/robin", dir),
				want: path.Join(dir, "robin"),
			}, {
				// Non existing file
				in:   fmt.Sprintf("%s/joker", dir),
				want: path.Join(dir, "joker"),
			}, {
				// Non existing-file, non minimal-file
				in:   fmt.Sprintf("%s/../robin", dir),
				want: path.Join(dir, "..", "robin"),
			}, {
				// Relative file
				in:   fmt.Sprintf("../../../../../../../../../../../../../../%s/joker", dir),
				want: path.Join(dir, "joker"),
			}, {
				// Non-existing directory
				in:   fmt.Sprintf("%s/batcave/joker", dir),
				want: "",
			}, {
				// Directory
				in:   fmt.Sprintf("%s/batman", dir),
				want: "",
			},
		}
		for _, c := range cases {
			got := FilePath("")
			err := got.UnmarshalText([]byte(c.in))
			if err != nil && c.want != "" {
				t.Errorf("UnmarshalText(%q) error:\n%+v", c.in, err)
				continue
			}
			if err == nil && c.want == "" {
				t.Errorf("UnmarshalText(%q) == %v but expected error", c.in, got)
			}
			if err == nil && c.want != string(got) {
				t.Errorf("UnmarshalText(%q) == %v but expected %v", c.in, got, c.want)
			}
		}
	})
}

func TestUnmarshalDirectoryPath(t *testing.T) {
	inPreparedDirectory(t, func(dir string) {
		cases := []struct {
			in   string
			want string
		}{
			{
				// Existing file
				in:   fmt.Sprintf("%s/robin", dir),
				want: "",
			}, {
				// Non existing directory
				in:   fmt.Sprintf("%s/joker", dir),
				want: path.Join(dir, "joker"),
			}, {
				// Non existing directory, non minimal directory
				in:   fmt.Sprintf("%s/../robin", dir),
				want: path.Join(dir, "..", "robin"),
			}, {
				// Relative file
				in:   fmt.Sprintf("../../../../../../../../../../../../../../%s/joker", dir),
				want: path.Join(dir, "joker"),
			}, {
				// Non existing directory
				in:   fmt.Sprintf("%s/batcave/joker", dir),
				want: "",
			}, {
				// Existing Directory
				in:   fmt.Sprintf("%s/batman", dir),
				want: path.Join(dir, "batman"),
			},
		}
		for _, c := range cases {
			got := DirectoryPath("")
			err := got.UnmarshalText([]byte(c.in))
			if err != nil && c.want != "" {
				t.Errorf("UnmarshalText(%q) error:\n%+v", c.in, err)
				continue
			}
			if err == nil && c.want == "" {
				t.Errorf("UnmarshalText(%q) == %v but expected error", c.in, got)
			}
			if err == nil && c.want != string(got) {
				t.Errorf("UnmarshalText(%q) == %v but expected %v", c.in, got, c.want)
			}
		}
	})
}

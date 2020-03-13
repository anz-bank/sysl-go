package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindConfigFilename(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "valid json",
			in:   "foo",
			out:  "foo.json",
		},
		{
			name: "valid yaml",
			in:   "bar",
			out:  "bar.yaml",
		},
		{
			name: "missing",
			in:   "missing",
			out:  "",
		},
	}

	dir, err := ioutil.TempDir("", "TestDirectoryConfigProvider")
	require.NoError(t, err, "error during test setup: failed to create temp dir")
	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			fmt.Printf("warning: failed to remove temp dir: %+v", err)
		}
	}()
	require.NoError(t, ioutil.WriteFile(filepath.Join(dir, "foo.json"), []byte(""), 0644))
	require.NoError(t, ioutil.WriteFile(filepath.Join(dir, "bar.yaml"), []byte(""), 0644))
	require.NoError(t, ioutil.WriteFile(filepath.Join(dir, "beef.yaml"), []byte(""), 0644))
	require.NoError(t, ioutil.WriteFile(filepath.Join(dir, "blarg.json"), []byte(""), 0644))

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			expected := filepath.Join(dir, tt.out)
			if tt.out == "" {
				expected = ""
			}
			require.Equal(t, expected, FindConfigFilename(dir, tt.in))
		})
	}
}

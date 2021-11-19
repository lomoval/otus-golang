package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	t.Run("main-test", func(t *testing.T) {
		expected := Environment{
			"BAR":   {Value: "bar", NeedRemove: false},
			"EMPTY": {Value: "", NeedRemove: false},
			"FOO":   {Value: "   foo\nwith new line", NeedRemove: false},
			"HELLO": {Value: "\"hello\"", NeedRemove: false},
			"UNSET": {Value: "", NeedRemove: true},
		}

		actual, err := ReadDir("./testdata/env")
		require.NoError(t, err)

		require.Equal(t, len(expected), len(actual))
		for s, value := range actual {
			e := expected[s]
			require.Equal(t, e.Value, value.Value)
			require.Equal(t, e.NeedRemove, value.NeedRemove)
		}
	})

	t.Run("right-trim", func(t *testing.T) {
		tmpDir := prepareTempDir(t)
		filepath := path.Join(tmpDir, "RIGHT_TRIM")
		creteFile(t, filepath, "test-data   \t ")

		actual, err := ReadDir(tmpDir)
		require.NoError(t, err)

		expected := Environment{
			"RIGHT_TRIM": {Value: "test-data", NeedRemove: false},
		}

		require.Equal(t, len(expected), len(actual))
		for s, value := range actual {
			e := expected[s]
			require.Equal(t, e.Value, value.Value)
			require.Equal(t, e.NeedRemove, value.NeedRemove)
		}
	})

	t.Run("non-exist-dir", func(t *testing.T) {
		_, err := ReadDir("_some_non_exist_dir_")
		require.Error(t, err)
	})

	t.Run("incorrect-file", func(t *testing.T) {
		tmpDir := prepareTempDir(t)
		filepath := path.Join(tmpDir, "NAME=NAME")
		creteFile(t, filepath, "incorrect")

		_, err := ReadDir(tmpDir)
		require.Error(t, err)
	})
}

func creteFile(t *testing.T, filepath string, data string) {
	t.Helper()

	f, err := os.Create(filepath)
	require.NoError(t, err, "failed to create temp file")
	defer func() {
		require.NoError(t, f.Close(), "failed to close temp file")
	}()

	_, err = f.Write([]byte(data))
	require.NoError(t, err, "failed write to temp file")
}

func prepareTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "envdir-tool-*")
	require.NoError(t, err, "failed to create temp dir")

	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir), "failed to remove temp dir")
	})
	return dir
}

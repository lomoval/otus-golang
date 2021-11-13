package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEdgeCopyCases(t *testing.T) {
	tempDir := prepareTempDir(t)
	testCases := []struct {
		Name         string
		FromFile     string
		Limit        int64
		Offset       int64
		ExpectedFile string
	}{
		{
			Name:         "EmptyFile",
			FromFile:     creteFileWithSize(t, tempDir, 0),
			Limit:        0,
			Offset:       0,
			ExpectedFile: creteFileWithSize(t, tempDir, 0),
		},
		{
			Name:         "OneByteFile",
			FromFile:     creteFileWithSize(t, tempDir, 1),
			Limit:        0,
			Offset:       0,
			ExpectedFile: creteFileWithSize(t, tempDir, 1),
		},
		{
			Name:         "EmptyByOffset",
			FromFile:     creteFileWithSize(t, tempDir, 1),
			Limit:        0,
			Offset:       1,
			ExpectedFile: creteFileWithSize(t, tempDir, 0),
		},
		{
			Name:         "FullWithOverLimit",
			FromFile:     creteFileWithSize(t, tempDir, 20),
			Limit:        200,
			Offset:       0,
			ExpectedFile: creteFileWithSize(t, tempDir, 20),
		},
		{
			Name:         "EmptyWithOffsetLimit",
			FromFile:     creteFileWithSize(t, tempDir, 20),
			Limit:        20,
			Offset:       20,
			ExpectedFile: creteFileWithSize(t, tempDir, 0),
		},
		{
			Name:         "OneLastByte",
			FromFile:     creteFileWithSize(t, tempDir, 20),
			Limit:        20,
			Offset:       19,
			ExpectedFile: creteFileWithLimitOffset(t, tempDir, 19, 1),
		},
		{
			Name:         "MiddleBytes",
			FromFile:     creteFileWithSize(t, tempDir, 20),
			Limit:        4,
			Offset:       8,
			ExpectedFile: creteFileWithLimitOffset(t, tempDir, 8, 4),
		},
		{
			Name:         "FullFileByLimit",
			FromFile:     creteFileWithSize(t, tempDir, 20),
			Limit:        20,
			Offset:       0,
			ExpectedFile: creteFileWithSize(t, tempDir, 20),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d - %s", i+1, testCase.Name), func(t *testing.T) {
			toFile := path.Join(tempDir, testCase.Name)
			require.NoError(t, Copy(testCase.FromFile, toFile, testCase.Offset, testCase.Limit))
			expected, err := os.ReadFile(testCase.ExpectedFile)
			require.NoError(t, err)
			actual, err := os.ReadFile(toFile)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})
	}
}

func TestIncorrectCopyCases(t *testing.T) {
	tempDir := prepareTempDir(t)
	testCases := []struct {
		Name     string
		ToFile   string
		FromFile string
		Limit    int64
		Offset   int64
		Error    error
	}{
		{
			Name:     "NonExistFromFile",
			ToFile:   "./non-exist-from-file",
			FromFile: "./non-exist-to-file",
			Limit:    0,
			Offset:   0,
			Error:    ErrOpenFile,
		},
		{
			Name:     "SameFiles",
			ToFile:   "./file",
			FromFile: "./file",
			Limit:    0,
			Offset:   0,
			Error:    ErrCopying,
		},
		{
			Name:     "OffsetExceedsFileSize",
			ToFile:   "./file",
			FromFile: creteFileWithSize(t, tempDir, 100),
			Limit:    0,
			Offset:   101,
			Error:    ErrOffsetExceedsFileSize,
		},
		{
			Name:     "IncorrectToFile",
			ToFile:   "/./!@#$%^&*()file",
			FromFile: creteFileWithSize(t, tempDir, 1),
			Limit:    0,
			Offset:   0,
			Error:    ErrOpenFile,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d - %s", i+1, testCase.Name), func(t *testing.T) {
			require.Equal(t, testCase.Error, Copy(testCase.FromFile, testCase.ToFile, testCase.Offset, testCase.Limit))
		})
	}
}

func creteFileWithSize(t *testing.T, dir string, size int64) string {
	t.Helper()
	return creteFileWithLimitOffset(t, dir, 0, size)
}

func creteFileWithLimitOffset(t *testing.T, dir string, offset int64, limit int64) string {
	t.Helper()
	b := make([]byte, limit)
	var val byte
	for i := int64(0); i < offset+limit; i++ {
		val++
		if i >= offset {
			b[i-offset] = val
		}
	}
	f, err := os.CreateTemp(dir, "")
	require.NoError(t, err, "failed to create temp file")
	defer func() {
		require.NoError(t, f.Close(), "failed to close temp file")
	}()

	_, err = f.Write(b)
	require.NoError(t, err, "failed write to temp file")
	return f.Name()
}

func prepareTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "file-copying-*")
	require.NoError(t, err, "failed to create temp dir")

	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

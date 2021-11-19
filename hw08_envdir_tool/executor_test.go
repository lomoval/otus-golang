package main

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("main-test", func(t *testing.T) {
		require.NoError(t, os.Setenv("ADDED", "from original env"))
		envs := Environment{
			"BAR":   {Value: "bar", NeedRemove: false},
			"EMPTY": {Value: "", NeedRemove: false},
			"FOO":   {Value: "   foo\nwith new line", NeedRemove: false},
			"HELLO": {Value: "\"hello\"", NeedRemove: false},
			"UNSET": {Value: "", NeedRemove: true},
		}
		expected := `HELLO is ("hello")
BAR is (bar)
FOO is (   foo
with new line)
UNSET is ()
ADDED is (from original env)
EMPTY is ()
arguments are arg1=1 arg2=2` + osEndLine

		checkExecutorOutput(t, []string{shell, path.Join("testdata", echoScript), "arg1=1", "arg2=2"}, envs, expected)
	})

	t.Run("no-environments", func(t *testing.T) {
		require.NoError(t, os.Unsetenv("ADDED"))
		require.NoError(t, os.Unsetenv("HELLO"))
		require.NoError(t, os.Unsetenv("BAR"))
		require.NoError(t, os.Unsetenv("FOO"))
		require.NoError(t, os.Unsetenv("EMPTY"))
		envs := Environment{}
		expected := `HELLO is ()
BAR is ()
FOO is ()
UNSET is ()
ADDED is ()
EMPTY is ()
arguments are arg1=1 arg2=2` + osEndLine

		checkExecutorOutput(t, []string{shell, path.Join("testdata", echoScript), "arg1=1", "arg2=2"}, envs, expected)
	})

	t.Run("no-params", func(t *testing.T) {
		require.Equal(t, -1, RunCmd([]string{}, Environment{}))
	})

	t.Run("incorrect-command", func(t *testing.T) {
		require.Equal(t, -1, RunCmd([]string{"__incorrect_command__"}, Environment{}))
	})

	t.Run("incorrect-command-parameter", func(t *testing.T) {
		require.Less(t, 0, RunCmd([]string{shell, "__incorrect_param__"}, Environment{}))
	})
}

func checkExecutorOutput(t *testing.T, args []string, envs Environment, expectedOutput string) {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	require.Equal(t, RunCmd(args, envs), 0)
	require.NoError(t, w.Close())

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)
		outC <- buf.String()
		close(outC)
	}()
	require.Equal(t, expectedOutput, <-outC)
}

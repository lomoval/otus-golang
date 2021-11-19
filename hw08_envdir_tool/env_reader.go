package main

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	envs := Environment{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, errors.New("failed to get info about file " + entry.Name())
		}
		if strings.Contains(entry.Name(), "=") {
			return nil, errors.New("incorrect file name " + entry.Name())
		}

		env := EnvValue{Value: "", NeedRemove: info.Size() == 0}
		if !env.NeedRemove {
			env.Value, err = readLine(path.Join(dir, entry.Name()))
			if err != nil {
				return nil, err
			}
			env.Value = normalizeValue(env.Value)
		}
		envs[entry.Name()] = env
	}
	return envs, nil
}

func normalizeValue(val string) string {
	val = strings.TrimRight(val, " \t")
	return string(bytes.ReplaceAll([]byte(val), []byte{byte('\x00')}, []byte{byte('\n')}))
}

func readLine(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return "", nil
}

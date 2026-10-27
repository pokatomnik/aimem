package init

import (
	"errors"
	"fmt"
	"os"
	"path"
)

type (
	confirmation interface {
		Confirm(prompt string) (bool, error)
	}

	Initializer struct {
		filePath              string
		defaultConfigContents string
		confirmation          confirmation
	}
)

func New(filePath, defaultConfigContents string, confirmation confirmation) *Initializer {
	return &Initializer{
		filePath:              filePath,
		defaultConfigContents: defaultConfigContents,
		confirmation:          confirmation,
	}
}

func (i *Initializer) dumpConfigFile() error {
	dir := path.Dir(i.filePath)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	return os.WriteFile(i.filePath, []byte(i.defaultConfigContents), 0644)
}

func (i *Initializer) confirmOverwrite() bool {
	res, err := i.confirmation.Confirm(
		fmt.Sprintf("Configuration already exists at %s. Overwrite?", i.filePath),
	)
	if err != nil {
		return false
	}

	return res
}

func (i *Initializer) Dump() (bool, error) {
	stat, err := os.Stat(i.filePath)

	if errors.Is(err, os.ErrNotExist) {
		err = i.dumpConfigFile()

		return err == nil, err
	}

	if err != nil {
		return false, err
	}

	if stat.IsDir() {
		return false, fmt.Errorf("failed to save configuration: target file is a directory")
	}

	overwrite := i.confirmOverwrite()
	if overwrite {
		err = i.dumpConfigFile()
		return err == nil, err
	}

	return false, nil
}

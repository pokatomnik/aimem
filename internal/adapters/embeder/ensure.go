package embeder

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

var modelFiles = []string{
	"config.json",
	"tokenizer.json",
	"config_sentence_transformers.json",
	"model.safetensors",
}

func ensureModel(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	for _, name := range modelFiles {
		path := filepath.Join(dir, name)

		info, err := os.Stat(path)
		if err == nil && info.Size() > 0 {
			continue
		}
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		if err := unpackFile(name, path); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}

func unpackFile(name string, dest string) error {
	weightsFile, err := gemmaDir.Open(path.Join(gemmaDirName, name))
	if err != nil {
		return err
	}
	defer weightsFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, weightsFile); err != nil {
		return fmt.Errorf("Cannot unpack weights file %s with error %w", name, err)
	}

	return nil
}

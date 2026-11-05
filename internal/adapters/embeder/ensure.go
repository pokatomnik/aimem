package embeder

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const modelURL = "https://huggingface.co/google/embeddinggemma-2/resolve/main/"

var modelFiles = []string{
	"config.json",
	"tokenizer.json",
	"config_sentence_transformers.json",
	"model.safetensors",
}

func ensureModel(ctx context.Context, dir string) error {
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

		if err := downloadFile(ctx, modelURL+name, path); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}

func downloadFile(
	ctx context.Context,
	url string,
	dest string,
) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, url, nil,
	)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %s", resp.Status)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dest), ".model-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), dest)
}

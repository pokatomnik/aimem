package kv

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/pokatomnik/gomem/internal/shared/message"
)

func (k *KV) appendPlainTextMessage(content string) error {
	filePath := path.Join(k.dataPath, "plain.jsonl")

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	msg := message.Message{Data: content}

	msgStr, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(file, "%s\n", msgStr); err != nil {
		return err
	}

	return nil
}

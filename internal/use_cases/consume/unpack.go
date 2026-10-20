package consume

import (
	"encoding/json"

	"github.com/pokatomnik/gomem/internal/shared/message"
)

func unpack(data []byte) (string, error) {
	var message message.Message

	err := json.Unmarshal(data, &message)
	if err != nil {
		return "", err
	}

	return message.Data, nil
}

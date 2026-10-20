package dump

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pokatomnik/gomem/internal/shared/message"
)

type Formatter struct{}

func (f *Formatter) Write(data []byte) (int, error) {
	factStr := string(data)
	fact := message.Message{
		Data: factStr,
	}

	factJSON, err := json.Marshal(fact)
	if err != nil {
		return 0, err
	}

	bData, err := fmt.Fprint(os.Stdout, string(factJSON))
	if err != nil {
		return 0, err
	}

	bSep, err := fmt.Fprintf(os.Stdout, "\n")
	if err != nil {
		return bData, err
	}

	return bData + bSep, nil
}

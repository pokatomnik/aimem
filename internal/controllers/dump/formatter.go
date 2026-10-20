package dump

import (
	"encoding/json"
	"fmt"
	"os"
)

type Formatter struct{}

func (f *Formatter) Write(data []byte) (int, error) {
	factStr := string(data)
	fact := BareFact{
		Fact: factStr,
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

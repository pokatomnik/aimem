package input

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/pokatomnik/gomem/internal/shared/colors"
)

type Input struct {
	buf *bufio.Reader
}

func New(input io.Reader) *Input {
	return &Input{
		buf: bufio.NewReader(input),
	}
}

func (i *Input) Receive() ([]byte, error) {
	data, err := i.buf.ReadBytes('\n')

	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	return data, err
}

func (i *Input) Confirm(prompt string) (bool, error) {
	colors.Primary().Print(prompt + " ")
	colors.Secondary().Print("[Y/n] ")

	response, err := i.Receive()
	if err != nil && !(errors.Is(err, io.EOF) && len(response) > 0) {
		return false, err
	}

	switch strings.ToLower(strings.TrimSpace(string(response))) {
	case "", "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid confirmation response %q: expected y or n", response)
	}
}

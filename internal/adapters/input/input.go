package input

import (
	"bufio"
	"io"
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

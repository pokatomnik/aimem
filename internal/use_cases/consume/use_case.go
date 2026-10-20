package consume

import (
	"context"
	"errors"
	"io"

	"github.com/pokatomnik/gomem/internal/entities/ollama"
)

type (
	receiver interface {
		Receive() ([]byte, error)
	}

	ai interface {
		GetEmbeddings(ctx context.Context, text string) (ollama.Embeddings, error)
		Summarize(ctx context.Context, text string) (string, error)
	}

	store interface {
		AddFact(ctx context.Context, vector []float32, summary, content string) error
	}

	Consumer struct {
		ai ai
		s  store
		r  receiver
	}
)

func New(ai ai, s store, r receiver) *Consumer {
	return &Consumer{ai, s, r}
}

func (d *Consumer) Consume(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// packed message
		data, err := d.r.Receive()
		if errors.Is(err, io.EOF) {
			if len(data) == 0 {
				return nil
			}

			fact, err := unpack(data)
			if err != nil {
				return err
			}

			return d.addFact(ctx, fact)
		}

		if err != nil {
			return err
		}

		fact, err := unpack(data)
		if err != nil {
			return err
		}

		err = d.addFact(ctx, fact)
		if err != nil {
			return err
		}
	}
}

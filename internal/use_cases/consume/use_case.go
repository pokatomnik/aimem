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

func (d *Consumer) Consume(ctx context.Context) []error {
	errs := make([]error, 0)

	for {
		if ctx.Err() != nil {
			break
		}

		// packed message
		data, err := d.r.Receive()
		if errors.Is(err, io.EOF) {
			if len(data) == 0 {
				break
			}

			fact, err := unpack(data)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			err = d.addFact(ctx, fact)
			if err != nil {
				errs = append(errs, err)
			}
			break
		}

		if err != nil {
			errs = append(errs, err)
			break
		}

		fact, err := unpack(data)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = d.addFact(ctx, fact)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	return errs
}

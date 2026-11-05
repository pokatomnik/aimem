package embeder

import (
	"context"

	"github.com/townsendmerino/goinfer/embeddinggemma2"
)

func (e *Embedder) getEncoder(ctx context.Context) (*embeddinggemma2.Encoder, error) {
	e.mx.Lock()
	defer e.mx.Unlock()

	if err := ensureModel(ctx, e.modelDir); err != nil {
		return nil, err
	}

	if e.initialized {
		return e.encoder, nil
	}

	encoder, err := embeddinggemma2.LoadEncoder(e.modelDir)
	if err != nil {
		return nil, err
	}

	e.initialized = true
	e.encoder = encoder

	return encoder, nil
}

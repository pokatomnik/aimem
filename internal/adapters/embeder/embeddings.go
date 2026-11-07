package embeder

import (
	"context"
	"time"

	"github.com/pokatomnik/gomem/internal/entities/embeddings"
)

const (
	periodicCheck = time.Second
)

func (e *Embedder) GetEmbeddings(ctx context.Context, text string) (embeddings.Embeddings, error) {
	e.encoderMx.Lock()
	defer e.encoderMx.Unlock()
	defer func() {
		now := time.Now()
		e.lastUsed = &now
	}()

	enc, err := e.getEncoder()
	if err != nil {
		e.logger.Error("get encoder failed", "text_length", len(text), "error", err)

		return embeddings.Embeddings{}, err
	}

	vec, err := enc.EmbedText(text)
	if err != nil {
		e.logger.Error("embed text failed", "text_length", len(text), "error", err)

		return nil, err
	}

	vecs := [][]float32{vec}

	return embeddings.Embeddings(vecs), err
}

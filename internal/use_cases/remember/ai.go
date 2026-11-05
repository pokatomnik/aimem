package remember

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/embeddings"
	"golang.org/x/sync/errgroup"
)

func (r *Remember) getEmbeddingsAndSummary(ctx context.Context, fact string) (string, embeddings.Embeddings, error) {
	eg, ctx := errgroup.WithContext(ctx)
	sCh := make(chan string, 1)
	eCh := make(chan embeddings.Embeddings, 1)

	eg.Go(func() error {
		summary, err := r.summarizer.Summarize(ctx, fact)
		if err != nil {
			return err
		}

		sCh <- summary

		return nil
	})

	eg.Go(func() error {
		emb, err := r.embedder.GetEmbeddings(ctx, fact)
		if err != nil {
			return err
		}

		eCh <- emb

		return nil
	})

	err := eg.Wait()
	if err != nil {
		return "", embeddings.Embeddings{}, err
	}

	return <-sCh, <-eCh, nil
}

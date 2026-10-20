package kv

import (
	"context"
	"io"

	"github.com/gorse-io/xvec"
)

func (k *KV) Dump(ctx context.Context, to io.Writer) error {
	iter, err := k.collection.CreateIterator(ctx, xvec.IteratorOptions{
		Projection: xvec.Projection{
			OutputFields: []string{keyFieldFact},
		},
	})
	if err != nil {
		return err
	}
	defer iter.Close()

	doc, err := iter.Next()

	for doc != nil && err == nil {
		content, ok := doc.Fields[keyFieldFact].(string)
		if !ok {
			continue
		}
		to.Write([]byte(content))

		doc, err = iter.Next()
	}

	return nil
}

package kv

import (
	"context"

	"github.com/gorse-io/xvec"
)

type (
	config interface {
		GetDimensions() uint32
	}

	// collection is the part of xvec.Collection this repository uses. It is
	// declared here, next to its only consumer, so a test can supply a fake:
	// tests are hermetic and may not open a real database on disk.
	collection interface {
		Insert(ctx context.Context, documents []xvec.Document) ([]xvec.WriteResult, error)
		Query(ctx context.Context, query xvec.VectorQuery) ([]xvec.Document, error)
		Fetch(ctx context.Context, primaryKeys []string, projection xvec.Projection) ([]*xvec.Document, error)
		Delete(ctx context.Context, primaryKeys []string) ([]xvec.WriteResult, error)
		Close() error
	}

	// collectionOpener resolves the collection. New leaves it nil, which selects
	// the xvec-backed implementation; a test sets it to inject a fake.
	collectionOpener func(ctx context.Context) (collection, error)

	KV struct {
		config config
		// collection *xvec.Collection
		dataPath string
		schema   xvec.CollectionSchema
		open     collectionOpener
	}
)

func New(config config, dataPath string) *KV {
	index := xvec.NewHNSWIndexParams(xvec.MetricTypeCosine)
	schema := xvec.NewCollectionSchema(
		"memory",
		xvec.NewField(keyFieldCreatedAt, xvec.DataTypeString),
		xvec.NewField(keyFieldSummary, xvec.DataTypeString),
		xvec.NewField(keyFieldFact, xvec.DataTypeString),
		xvec.FieldSchema{
			Name:      keyFieldEmbedding,
			DataType:  xvec.DataTypeVectorFP32,
			Dimension: config.GetDimensions(),
			Index:     index,
		},
	)

	return &KV{
		config:   config,
		dataPath: dataPath,
		schema:   schema,
	}
}

func (kv *KV) getCollection(ctx context.Context) (collection, error) {
	if kv.open != nil {
		return kv.open(ctx)
	}

	return kv.openXvec(ctx)
}

func (kv *KV) openXvec(ctx context.Context) (collection, error) {
	options := xvec.NewCollectionOptions()

	opened, err := xvec.Open(ctx, kv.dataPath, options)
	if err == nil {
		return opened, nil
	}

	opened, err = xvec.CreateAndOpen(ctx, kv.dataPath, kv.schema, options)
	if err == nil {
		return opened, nil
	}

	return nil, err
}

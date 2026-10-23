package kv

import (
	"context"

	"github.com/gorse-io/xvec"
)

type (
	config interface {
		GetDimensions() uint32
	}

	KV struct {
		config config
		// collection *xvec.Collection
		dataPath string
		schema   xvec.CollectionSchema
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

func (kv *KV) getCollection(ctx context.Context) (*xvec.Collection, error) {

	options := xvec.NewCollectionOptions()

	collection, err := xvec.Open(ctx, kv.dataPath, options)
	if err == nil {
		return collection, nil
	}

	collection, err = xvec.CreateAndOpen(ctx, kv.dataPath, kv.schema, options)
	if err == nil {
		return collection, nil
	}

	return nil, err
}

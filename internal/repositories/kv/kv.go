package kv

import (
	"context"
	"os"
	"path"

	"github.com/gorse-io/xvec"
)

type (
	config interface {
		GetDimensions() uint32
	}

	KV struct {
		config     config
		collection *xvec.Collection
	}
)

func Open(ctx context.Context, config config) (*KV, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dataPath := path.Join(userHome, ".aimem")

	index := xvec.NewHNSWIndexParams(xvec.MetricTypeCosine)
	schema := xvec.NewCollectionSchema(
		"memory",
		xvec.NewField(keyFieldSummary, xvec.DataTypeString),
		xvec.NewField(keyFieldFact, xvec.DataTypeString),
		xvec.FieldSchema{
			Name:      keyFieldEmbedding,
			DataType:  xvec.DataTypeVectorFP32,
			Dimension: config.GetDimensions(),
			Index:     index,
		},
	)

	options := xvec.NewCollectionOptions()

	collection, err := xvec.Open(ctx, dataPath, options)
	if err == nil {
		return &KV{config, collection}, nil
	}

	collection, err = xvec.CreateAndOpen(ctx, dataPath, schema, options)
	if err == nil {
		return &KV{config, collection}, nil
	}

	return nil, err
}

func (k *KV) Close() error {
	return k.collection.Close()
}

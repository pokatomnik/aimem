package qdrant

import (
	"context"
	"errors"
	"net/http"
)

type (
	// config is the part of the configuration file this repository needs. It is
	// declared here, next to its only consumer, so the package does not depend on
	// the config package.
	config interface {
		// GetDimensions is the vector size the collection is created with. It must
		// match the embedding model, e.g. 768 for embeddinggemma-2:740m. Changing it
		// after the collection exists makes Qdrant reject every insert.
		GetDimensions() uint32
		// GetQdrantHost is the base URL of the server, without a trailing slash,
		// e.g. "http://127.0.0.1:6333". It must not be empty: without a host every
		// request fails before it is sent.
		GetQdrantHost() string
		// GetQdrantAPIKey is sent in the api-key header. Return "" when the server
		// runs without authentication, which is the docker-compose default: the
		// header is then omitted entirely and the request is accepted.
		GetQdrantAPIKey() string
	}

	Qdrant struct {
		config config
		client *http.Client
	}
)

// New builds the repository. A nil client falls back to http.DefaultClient,
// which matches how the rest of the application makes requests: without a
// deadline, cancelled only through the context.
func New(config config, client *http.Client) *Qdrant {
	if client == nil {
		client = http.DefaultClient
	}

	return &Qdrant{
		config: config,
		client: client,
	}
}

// ensureCollection creates the user's collection when it is missing, mirroring
// kv.getCollection, which opens the store and creates it on first use. Each user
// gets their own, so the first write of a new user also provisions their storage.
func (q *Qdrant) ensureCollection(ctx context.Context, userID string) error {
	var existence collectionExistence

	if err := q.do(ctx, http.MethodGet, pathCollectionExists(userID), nil, nil, &existence); err != nil {
		return err
	}

	if existence.Exists {
		return nil
	}

	return q.createCollection(ctx, userID)
}

func (q *Qdrant) createCollection(ctx context.Context, userID string) error {
	request := createCollectionRequest{
		Vectors: map[string]vectorParams{
			keyFieldEmbedding: {
				Size:     q.config.GetDimensions(),
				Distance: distanceCosine,
				Datatype: datatypeFloat32,
			},
		},
	}

	err := q.do(ctx, http.MethodPut, pathCollection(userID), nil, request, nil)

	// The MCP HTTP and web servers answer requests concurrently, so two of them
	// can decide to create the collection at the same moment. Qdrant rejects the
	// loser with 409; the collection it wanted now exists either way.
	var conflict *httpError
	if errors.As(err, &conflict) && conflict.StatusCode == http.StatusConflict {
		return nil
	}

	return err
}

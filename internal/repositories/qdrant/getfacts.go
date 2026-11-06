package qdrant

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pokatomnik/gomem/internal/entities/fact"
)

// GetFactsList returns the most similar facts, ordered by similarity.
//
// The query runs against the caller's collection, so a user's recall only ever
// sees their own facts. The ranking, the similarity threshold and topK apply
// within that collection, not across the server.
//
// similarityThreshold carries the same meaning as in kv — a floor on cosine
// similarity, where 0 matches everything and 1 requires an identical vector — but
// it is compared against Qdrant's score directly. kv stores cosine as a distance
// and flips it with "1 - result.Score"; Qdrant reports a similarity already, so
// applying that inversion here would compare the wrong number and drop the
// closest facts.
func (q *Qdrant) GetFactsList(ctx context.Context, userID string, vector []float32, topK int, similarityThreshold float32) (facts []fact.FactListView, err error) {
	// xvec rejects a non-positive TopK too; Qdrant's limit has the same floor.
	if topK <= 0 {
		err := fmt.Errorf("topK must be positive, got %d", topK)
		q.logger.Warn("topK must be positive", "user_id", userID, "top_k", topK, "error", err)

		return nil, err
	}

	if err := q.ensureCollection(ctx, userID); err != nil {
		q.logger.Error("ensure collection failed", "user_id", userID, "error", err)

		return []fact.FactListView{}, err
	}

	request := queryPointsRequest{
		Query:       vector,
		Using:       keyFieldEmbedding,
		Limit:       topK,
		WithPayload: []string{keyFieldCreatedAt, keyFieldSummary},
		WithVector:  false,
	}

	var response queryPointsResponse

	if err := q.do(ctx, http.MethodPost, pathPointsQuery(userID), nil, request, &response); err != nil {
		q.logger.Error("query facts failed", "user_id", userID, "top_k", topK, "threshold", similarityThreshold, "error", err)

		return nil, err
	}

	facts = make([]fact.FactListView, 0, len(response.Points))

	for _, point := range response.Points {
		if point.Score < similarityThreshold {
			continue
		}

		createdAt, err := stringField(point.Payload, keyFieldCreatedAt)
		if err != nil {
			q.logger.Error("read fact created_at failed", "user_id", userID, "fact_id", point.ID.UUID, "field", keyFieldCreatedAt, "error", err)

			return nil, err
		}

		summary, err := stringField(point.Payload, keyFieldSummary)
		if err != nil {
			q.logger.Error("read fact summary failed", "user_id", userID, "fact_id", point.ID.UUID, "field", keyFieldSummary, "error", err)

			return nil, err
		}

		facts = append(facts, fact.FactListView{
			ID:        point.ID.UUID,
			CreatedAt: createdAt,
			Summary:   summary,
		})
	}

	return facts, nil
}

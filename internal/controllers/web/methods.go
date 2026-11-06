package web

import (
	"context"
	"time"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

// Methods holds the handlers of the generated server.
//
// authRequired decides whether a request needs a resolved user. The generated
// server treats the bearer token as optional, because whether the header is
// mandatory depends on the storage backend and only the running configuration
// knows that; this field carries the answer from startup into the handlers.
type Methods struct {
	remember     Remember
	recall       Recall
	getFact      GetFact
	removeFact   RemoveFact
	authRequired bool
	logger       logger
}

// getUserID returns the user the request was authenticated as.
//
// With authentication off the context is never read: the embedded backend ignores
// userID entirely, so there is nothing to look up and no header that could have
// put anything there. With authentication on an empty ID is refused, because
// falling back to the base collection would silently mix every user's facts into
// one.
func (m *Methods) getUserID(ctx context.Context) (string, error) {
	if !m.authRequired {
		return "", nil
	}

	id, _ := ctx.Value(userIDKey{}).(string)
	if id == "" {
		m.logger.Warn("request has no authenticated user", "error", apperrors.ErrUserNotFound)

		return "", apperrors.ErrUserNotFound
	}

	return id, nil
}

func (m *Methods) AddFact(ctx context.Context, req *server.AddFactReq) (server.AddFactRes, error) {
	userID, err := m.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	summary, err := m.remember.Remember(ctx, userID, req.Value.Fact)
	if err != nil {
		m.logger.Error("remember failed", "user_id", userID, "fact_length", len(req.Value.Fact), "error", err)

		return nil, err
	}

	return &server.AddFactOK{
		Value: server.FactAddedOK{
			Summary: summary,
		},
	}, nil
}

func (m *Methods) GetFactByID(ctx context.Context, params server.GetFactByIDParams) (server.GetFactByIDRes, error) {
	userID, err := m.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	res, err := m.getFact.GetFact(ctx, userID, params.FactID)
	if err != nil {
		m.logger.Error("get fact failed", "user_id", userID, "fact_id", params.FactID, "error", err)

		return nil, err
	}

	createdAt, err := time.Parse(time.RFC3339, res.CreatedAt)
	if err != nil {
		m.logger.Error("parse fact created_at failed", "fact_id", params.FactID, "created_at", res.CreatedAt, "error", err)

		return nil, err
	}

	return &server.GetFactByIDOK{
		Value: server.Fact{
			ID:        res.ID,
			CreatedAt: createdAt,
			Summary:   res.Summary,
			Content:   res.Content,
		},
	}, nil
}

func (m *Methods) Search(ctx context.Context, req *server.SearchRequestBody) (server.SearchRes, error) {
	userID, err := m.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	res, err := m.recall.Recall(ctx, userID, req.Query)
	if err != nil {
		m.logger.Error("recall failed", "user_id", userID, "query_length", len(req.Query), "error", err)

		return nil, err
	}

	found := make([]server.SearchResultItem, 0, len(res))
	for _, f := range res {
		createdAt, err := time.Parse(time.RFC3339, f.CreatedAt)
		if err != nil {
			m.logger.Error("parse search result created_at failed", "fact_id", f.ID, "created_at", f.CreatedAt, "error", err)

			return nil, err
		}
		found = append(found, server.SearchResultItem{
			ID:        f.ID,
			CreatedAt: createdAt,
			Summary:   f.Summary,
		})
	}

	return &server.SearchOK{
		Value: server.SearchResult{
			Found: found,
		},
	}, nil
}

func (m *Methods) RemoveFactByID(ctx context.Context, params server.RemoveFactByIDParams) (server.RemoveFactByIDRes, error) {
	userID, err := m.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = m.removeFact.RemoveFactByID(ctx, userID, params.FactID)
	if err != nil {
		m.logger.Error("remove fact failed", "user_id", userID, "fact_id", params.FactID, "error", err)

		return nil, err
	}

	return &server.RemoveFactByIDOK{
		Value: server.Nothing{},
	}, nil
}

package web

import (
	"context"
	"time"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

type Methods struct {
	remember   Remember
	recall     Recall
	getFact    GetFact
	removeFact RemoveFact
}

func (m *Methods) getUserID(ctx context.Context) (string, error) {
	id, ok := ctx.Value(userIDKey{}).(string)
	if !ok || id == "" {
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
		return nil, err
	}

	createdAt, err := time.Parse(time.RFC3339, res.CreatedAt)
	if err != nil {
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
		return nil, err
	}

	found := make([]server.SearchResultItem, 0, len(res))
	for _, f := range res {
		createdAt, err := time.Parse(time.RFC3339, f.CreatedAt)
		if err != nil {
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
		return nil, err
	}

	return &server.RemoveFactByIDOK{
		Value: server.Nothing{},
	}, nil
}

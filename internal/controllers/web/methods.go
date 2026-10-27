package web

import (
	"context"
	"time"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
)

type Methods struct {
	remember   Remember
	recall     Recall
	getFact    GetFact
	removeFact RemoveFact
}

func (m *Methods) AddFact(ctx context.Context, req *server.AddFactReq) (server.AddFactRes, error) {
	summary, err := m.remember.Remember(ctx, req.Value.Fact)
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
	res, err := m.getFact.GetFact(ctx, params.FactID)
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
	res, err := m.recall.Recall(ctx, req.Query)
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

func (m *Methods) RemoveFactByID(ctx context.Context, params server.RemoveFactByIDParams) (*server.RemoveFactByIDOK, error) {
	err := m.removeFact.RemoveFactByID(ctx, params.FactID)
	if err != nil {
		return nil, err
	}

	return &server.RemoveFactByIDOK{
		Value: server.Nothing{},
	}, nil
}

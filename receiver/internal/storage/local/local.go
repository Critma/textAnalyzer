package local

import (
	"receiver/internal/storage"

	"github.com/google/uuid"
)

func New() storage.Store {
	return storage.Store{
		Requests: &RequestStore{},
	}
}

var store []*storage.TextRequest = make([]*storage.TextRequest, 0)

type RequestStore struct {
}

func (s *RequestStore) CreateRequest(text string) (*storage.TextRequest, error) {
	if text == "" {
		return nil, storage.ErrEmptyText
	}

	request := &storage.TextRequest{
		ID:     uuid.New(),
		Text:   text,
		Status: storage.InProcess,
	}
	store = append(store, request)
	return request, nil
}

func (s *RequestStore) UpdateRequest(id uuid.UUID, status storage.Status, analyze storage.AnalyzeResult) (*storage.TextRequest, error) {
	for _, request := range store {
		if request.ID == id {
			request.Status = status
			request.Analyze = analyze
			return request, nil
		}
	}
	return nil, storage.ErrNotFound
}

func (s *RequestStore) GetRequest(id uuid.UUID) (*storage.TextRequest, error) {
	for _, request := range store {
		if request.ID == id {
			return request, nil
		}
	}
	return nil, storage.ErrNotFound
}

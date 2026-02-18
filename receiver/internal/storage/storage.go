package storage

import (
	"errors"

	"github.com/google/uuid"
)

type Store struct {
	Requests interface {
		CreateRequest(text string) (*TextRequest, error)
		UpdateRequest(id uuid.UUID, status Status, analyze AnalyzeResult) (*TextRequest, error)
		GetRequest(id uuid.UUID) (*TextRequest, error)
	}
}

var (
	ErrEmptyText = errors.New("empty text")
	ErrNotFound  = errors.New("request not found")
)

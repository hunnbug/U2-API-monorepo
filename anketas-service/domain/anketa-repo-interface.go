package domain

import (
	"context"

	"github.com/google/uuid"
)

type AnketaRepository interface {
	Create(ctx context.Context, anketa Anketa) error
	Update(ctx context.Context, id uuid.UUID, update map[string]any) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (Anketa, error)
	GetAnketas(ctx context.Context, pref PreferredAnketaGender, limit int) ([]Anketa, error)
}

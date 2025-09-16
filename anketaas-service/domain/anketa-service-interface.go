package domain

import (
	"context"

	"github.com/google/uuid"
)

type AnketaService interface {
	Create(
		ctx context.Context,
		username string,
		gender string,
		preferredGender string,
		description string,
		tags []string,
		photos []string,
	) error
	GetAnketaByID(ctx context.Context, id uuid.UUID) (Anketa, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, update map[string]any) error
	GetAnketas(ctx context.Context, pref PreferredAnketaGender, limit int) ([]Anketa, error)
}

package domain

import "github.com/google/uuid"

type AnketaRepository interface {
	Create(anketa Anketa) error
	Update(id uuid.UUID, update map[string]any) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (Anketa, error)
}

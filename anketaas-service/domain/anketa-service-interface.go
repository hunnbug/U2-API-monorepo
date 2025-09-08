package domain

import "github.com/google/uuid"

type AnketaService interface {
	Create(
		username string,
		gender string,
		preferredGender string,
		description string,
		tags []string,
		photos []string,
	) error
	GetAnketaByID(id uuid.UUID) (Anketa, error)
	Delete(id uuid.UUID) error
	Update(anketa *Anketa) error
}

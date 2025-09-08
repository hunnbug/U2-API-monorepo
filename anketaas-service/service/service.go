package service

import "anketas-service/anketaas-service/domain"

type AnketaService struct {
	repo domain.AnketaRepository
}

func NewAnketaService(repo domain.AnketaRepository) AnketaService {
	return AnketaService{repo}
}

package app

import "context"

type Service struct {
	ctx context.Context
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) startup(ctx context.Context) {
	s.ctx = ctx
}
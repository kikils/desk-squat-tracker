package service

import (
	"context"
	"time"

	"github.com/kikils/desk-squat-tracker/internal/usecase"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type StatsService struct {
	InputPort usecase.GetStatsInputPort

	ctx context.Context
}

func (s *StatsService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.ctx = ctx
	return nil
}

func (s *StatsService) GetStats(t time.Time) (*usecase.GetStatsOutput, error) {
	out, err := s.InputPort.Execute(s.ctx, t)
	if err != nil {
		return nil, err
	}
	return out, nil
}

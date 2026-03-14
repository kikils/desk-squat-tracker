package usecase

import (
	"context"
	"time"

	"cloud.google.com/go/civil"
	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
)

type GetStatsInputPort interface {
	Execute(ctx context.Context, t time.Time) (*GetStatsOutput, error)
}

type GetStatsOutput struct {
	RepCount int
}

type GetStatsInteractor struct {
	JudgementRepository repository.JudgementRepository
}

func NewGetStatsUsecase(judgementRepository repository.JudgementRepository) GetStatsInputPort {
	return &GetStatsInteractor{
		JudgementRepository: judgementRepository,
	}
}

func (i *GetStatsInteractor) Execute(ctx context.Context, t time.Time) (*GetStatsOutput, error) {
	repCount, err := i.JudgementRepository.CountRepsByDate(civil.DateOf(t.Local()))
	if err != nil {
		return nil, err
	}
	return &GetStatsOutput{
		RepCount: repCount,
	}, nil
}

package usecase

import (
	"context"
	"log"
	"time"

	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
	"github.com/kikils/desk-squat-tracker/internal/domain/service"
	"github.com/kikils/desk-squat-tracker/internal/errors"
)

// WatchSquatOutput は Execute の戻り値。顔検出時のみ Face / Judgement が非 nil。
type WatchSquatOutput struct {
	Face      *entity.Face
	Judgement *entity.Judgement
}

type WatchSquatInputPort interface {
	Execute(ctx context.Context, frame []byte, t time.Time) (*WatchSquatOutput, error)
}

type WatchSquatInteractor struct {
	FaceRepository      repository.FaceRepository
	JudgementRepository repository.JudgementRepository
	SquatJudger         service.SquatJudger
}

func NewWatchSquatUsecase(faceRepository repository.FaceRepository, judgementRepository repository.JudgementRepository, squatJudger service.SquatJudger) WatchSquatInputPort {
	return &WatchSquatInteractor{
		FaceRepository:      faceRepository,
		JudgementRepository: judgementRepository,
		SquatJudger:         squatJudger,
	}
}

func (i *WatchSquatInteractor) Execute(ctx context.Context, frame []byte, t time.Time) (*WatchSquatOutput, error) {
	face, err := i.FaceRepository.Detect(ctx, frame, t)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return &WatchSquatOutput{}, nil
		}
		return nil, err
	}

	judgement, err := i.SquatJudger.Judge(face)
	if err != nil {
		return nil, err
	}

	if judgement.IsRepCompleted {
		log.Println("rep completed")
	}

	if err := i.JudgementRepository.Save(judgement); err != nil {
		return nil, err
	}

	return &WatchSquatOutput{
		Face:      face,
		Judgement: judgement,
	}, nil
}

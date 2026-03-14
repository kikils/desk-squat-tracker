package usecase

import (
	"context"
	"fmt"

	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
)

type UpdateSettingInputPort interface {
	Execute(ctx context.Context, topRatio, bottomRatio float64) error
}

type UpdateSettingInteractor struct {
	SettingRepository repository.SettingRepository
}

func NewUpdateSettingUsecase(settingRepository repository.SettingRepository) UpdateSettingInputPort {
	return &UpdateSettingInteractor{
		SettingRepository: settingRepository,
	}
}

func (i *UpdateSettingInteractor) Execute(ctx context.Context, topRatio, bottomRatio float64) error {
	if topRatio <= 0 || topRatio >= 1 {
		return fmt.Errorf("topRatio must be in (0, 1), got %f", topRatio)
	}
	if bottomRatio <= 0 || bottomRatio >= 1 {
		return fmt.Errorf("bottomRatio must be in (0, 1), got %f", bottomRatio)
	}
	if bottomRatio >= topRatio {
		return fmt.Errorf("bottomRatio must be less than topRatio (bottomRatio=%f, topRatio=%f)", bottomRatio, topRatio)
	}
	return i.SettingRepository.Save(&entity.Setting{
		TopRatio:    topRatio,
		BottomRatio: bottomRatio,
	})
}

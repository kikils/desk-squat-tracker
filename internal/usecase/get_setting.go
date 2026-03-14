package usecase

import (
	"context"

	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
)

type GetSettingInputPort interface {
	Execute(ctx context.Context) (*GetSettingOutput, error)
}

type GetSettingOutput struct {
	TopRatio    float64
	BottomRatio float64
}

type GetSettingInteractor struct {
	SettingRepository repository.SettingRepository
}

func NewGetSettingUsecase(settingRepository repository.SettingRepository) GetSettingInputPort {
	return &GetSettingInteractor{
		SettingRepository: settingRepository,
	}
}

func (i *GetSettingInteractor) Execute(ctx context.Context) (*GetSettingOutput, error) {
	setting, err := i.SettingRepository.Get()
	if err != nil {
		return nil, err
	}
	return &GetSettingOutput{
		TopRatio:    setting.TopRatio,
		BottomRatio: setting.BottomRatio,
	}, nil
}

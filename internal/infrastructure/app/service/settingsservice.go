package service

import (
	"context"

	"github.com/kikils/desk-squat-tracker/internal/usecase"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type SettingsService struct {
	GetSettingInputPort    usecase.GetSettingInputPort
	UpdateSettingInputPort usecase.UpdateSettingInputPort

	ctx context.Context
}

func (s *SettingsService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.ctx = ctx
	return nil
}

func (s *SettingsService) GetSetting() (*usecase.GetSettingOutput, error) {
	return s.GetSettingInputPort.Execute(s.ctx)
}

func (s *SettingsService) UpdateSetting(topRatio, bottomRatio float64) error {
	return s.UpdateSettingInputPort.Execute(s.ctx, topRatio, bottomRatio)
}

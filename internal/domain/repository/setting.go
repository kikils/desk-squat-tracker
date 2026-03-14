package repository

import (
	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
)

type SettingRepository interface {
	Get() (*entity.Setting, error)
	Save(setting *entity.Setting) error
}

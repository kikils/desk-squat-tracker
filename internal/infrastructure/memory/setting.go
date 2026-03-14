package memory

import (
	"sync"

	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
)

type SettingRepository struct {
	setting *entity.Setting
	mu      sync.Mutex
}

func NewSettingRepository() repository.SettingRepository {
	return &SettingRepository{
		setting: entity.DefaultSetting(),
	}
}

func (r *SettingRepository) Get() (*entity.Setting, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// コピーを返す
	return &entity.Setting{
		TopRatio:    r.setting.TopRatio,
		BottomRatio: r.setting.BottomRatio,
	}, nil
}

func (r *SettingRepository) Save(s *entity.Setting) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.setting = &entity.Setting{
		TopRatio:    s.TopRatio,
		BottomRatio: s.BottomRatio,
	}
	return nil
}

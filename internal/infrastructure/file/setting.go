package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/kikils/desk-squat-tracker/internal/config"
	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
)

const settingsFilename = "settings.json"

type SettingRepository struct {
	mu   sync.Mutex
	path string
}

func NewSettingRepository() (repository.SettingRepository, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(userConfigDir, config.AppName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	return &SettingRepository{
		path: filepath.Join(dir, settingsFilename),
	}, nil
}

func (r *SettingRepository) Get() (*entity.Setting, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return entity.DefaultSetting(), nil
		}
		return nil, err
	}

	var s entity.Setting
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	// 不正・欠損値はデフォルトで補完（squat_judger が 0 や topRatio <= bottomRatio で壊れるのを防ぐ）
	def := entity.DefaultSetting()
	if s.TopRatio <= 0 || s.TopRatio >= 1 {
		s.TopRatio = def.TopRatio
	}
	if s.BottomRatio <= 0 || s.BottomRatio >= 1 {
		s.BottomRatio = def.BottomRatio
	}
	if s.BottomRatio >= s.TopRatio {
		s.TopRatio = def.TopRatio
		s.BottomRatio = def.BottomRatio
	}
	return &s, nil
}

func (r *SettingRepository) Save(s *entity.Setting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.path, data, 0600)
}

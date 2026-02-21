package repository

import (
	"time"

	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
)

type FaceRepository interface {
	Detect(frame []byte, t time.Time) (*entity.Face, error)
}

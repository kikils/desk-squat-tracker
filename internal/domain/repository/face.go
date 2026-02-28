package repository

import (
	"context"
	"time"

	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
)

type FaceRepository interface {
	Detect(ctx context.Context, frame []byte, t time.Time) (*entity.Face, error)
}

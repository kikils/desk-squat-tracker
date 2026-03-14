package repository

import (
	"cloud.google.com/go/civil"
	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
)

type JudgementRepository interface {
	Save(judgement *entity.Judgement) error
	GetLast() (*entity.Judgement, error)
	CountRepsByDate(date civil.Date) (int, error)
}

package repository

import "github.com/kikils/desk-squat-tracker/internal/domain/entity"

type JudgementRepository interface {
	Save(judgement *entity.Judgement) error
	GetLast() (*entity.Judgement, error)
}

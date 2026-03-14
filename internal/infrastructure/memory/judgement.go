package memory

import (
	"sync"

	"cloud.google.com/go/civil"
	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
	"github.com/kikils/desk-squat-tracker/internal/errors"
)

type JudgementRepository struct {
	judgements entity.Judgements
	mu         sync.Mutex
}

func NewJudgementRepository() repository.JudgementRepository {
	return &JudgementRepository{
		judgements: make(entity.Judgements, 0),
	}
}

func (r *JudgementRepository) Save(judgement *entity.Judgement) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.judgements = append(r.judgements, judgement)
	return nil
}

func (r *JudgementRepository) GetLast() (*entity.Judgement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.judgements) == 0 {
		return nil, errors.ErrNotFound.Errorf("not found last judgement")
	}
	return r.judgements[len(r.judgements)-1], nil
}

func (r *JudgementRepository) CountRepsByDate(date civil.Date) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, judgement := range r.judgements {
		if judgement.Date() == date && judgement.IsRepCompleted {
			count++
		}
	}
	return count, nil
}

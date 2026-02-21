package service

import (
	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
	"github.com/kikils/desk-squat-tracker/internal/errors"
)

type SquatJudger interface {
	Judge(face *entity.Face) (*entity.Judgement, error)
}

type squatJudgerImpl struct {
	FaceRepository      repository.FaceRepository
	JudgementRepository repository.JudgementRepository
}

func NewSquatJudger(faceRepository repository.FaceRepository, judgementRepository repository.JudgementRepository) SquatJudger {
	return &squatJudgerImpl{
		FaceRepository:      faceRepository,
		JudgementRepository: judgementRepository,
	}
}

func (s *squatJudgerImpl) Judge(face *entity.Face) (*entity.Judgement, error) {
	last, err := s.JudgementRepository.GetLast()
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, err
	}
	var prevState entity.DetectState
	if last != nil {
		prevState = last.State
	}

	// 顔の Y 位置（フレーム全体に対する比率）を算出
	centerY := face.CenterY()
	ratio := float64(centerY) / float64(face.FrameHeight)

	judgement := entity.NewJudgement(face)

	switch prevState {
	case entity.DetectStateUnknown, entity.DetectStateStanding:
		// 立位 or 未判定状態から、一定以上下がったら「しゃがみ始め」
		if ratio >= entity.DefaultTopRatio {
			judgement.State = entity.DetectStateGoingDown
		} else {
			judgement.State = entity.DetectStateStanding
		}

	case entity.DetectStateGoingDown:
		// さらに下がって十分な深さになったらボトム
		if ratio >= entity.DefaultTopRatio {
			judgement.State = entity.DetectStateBottom
		} else if ratio <= entity.DefaultBottomRatio {
			// 途中でまた上がり過ぎた場合は立位に戻す
			judgement.State = entity.DetectStateStanding
		} else {
			judgement.State = entity.DetectStateGoingDown
		}

	case entity.DetectStateBottom:
		// ボトムから上方向に戻り始めたら「立ち上がり」
		if ratio <= entity.DefaultBottomRatio {
			judgement.State = entity.DetectStateGoingUp
		} else {
			judgement.State = entity.DetectStateBottom
		}

	case entity.DetectStateGoingUp:
		// 十分に上がりきったら「立位」へ → 1 rep 完了
		if ratio <= entity.DefaultBottomRatio {
			judgement.State = entity.DetectStateStanding
			judgement.IsRepCompleted = true
		} else if ratio >= entity.DefaultTopRatio {
			// 再度下がり始めた場合は再度「しゃがみ始め」
			judgement.State = entity.DetectStateGoingDown
		} else {
			judgement.State = entity.DetectStateGoingUp
		}
	}

	return judgement, nil
}

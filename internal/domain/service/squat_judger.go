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
	SettingRepository   repository.SettingRepository
}

func NewSquatJudger(faceRepository repository.FaceRepository, judgementRepository repository.JudgementRepository, settingRepository repository.SettingRepository) SquatJudger {
	return &squatJudgerImpl{
		FaceRepository:      faceRepository,
		JudgementRepository: judgementRepository,
		SettingRepository:   settingRepository,
	}
}

func (s *squatJudgerImpl) Judge(face *entity.Face) (*entity.Judgement, error) {
	setting, err := s.SettingRepository.Get()
	if err != nil {
		return nil, err
	}
	topRatio := setting.TopRatio
	bottomRatio := setting.BottomRatio

	last, err := s.JudgementRepository.GetLast()
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, err
	}
	var prevState entity.DetectState
	if last != nil {
		prevState = last.State
	}

	// 顔の上端の Y 位置（フレーム全体に対する比率）を算出
	topY := face.TopY()
	ratio := float64(topY) / float64(face.FrameHeight)

	judgement := entity.NewJudgement(face)

	switch prevState {
	case entity.DetectStateUnknown, entity.DetectStateStanding:
		// 立位 or 未判定状態から、一定以上下がったら「しゃがみ始め」
		if ratio >= topRatio {
			judgement.State = entity.DetectStateGoingDown
		} else {
			judgement.State = entity.DetectStateStanding
		}

	case entity.DetectStateGoingDown:
		// さらに下がって十分な深さになったらボトム
		if ratio >= topRatio {
			judgement.State = entity.DetectStateBottom
		} else if ratio <= bottomRatio {
			// 途中でまた上がり過ぎた場合は立位に戻す
			judgement.State = entity.DetectStateStanding
		} else {
			judgement.State = entity.DetectStateGoingDown
		}

	case entity.DetectStateBottom:
		// ボトムから上方向に戻り始めたら「立ち上がり」
		if ratio <= bottomRatio {
			judgement.State = entity.DetectStateGoingUp
		} else {
			judgement.State = entity.DetectStateBottom
		}

	case entity.DetectStateGoingUp:
		// 十分に上がりきったら「立位」へ → 1 rep 完了
		if ratio <= bottomRatio {
			judgement.State = entity.DetectStateStanding
			judgement.IsRepCompleted = true
		} else if ratio >= topRatio {
			// 再度下がり始めた場合は再度「しゃがみ始め」
			judgement.State = entity.DetectStateGoingDown
		} else {
			judgement.State = entity.DetectStateGoingUp
		}
	}

	return judgement, nil
}

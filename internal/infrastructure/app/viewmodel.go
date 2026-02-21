package app

import (
	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	"github.com/kikils/desk-squat-tracker/internal/usecase"
)

var detectStateLabels = map[entity.DetectState]string{
	entity.DetectStateUnknown:   "unknown",
	entity.DetectStateStanding:  "standing",
	entity.DetectStateGoingDown: "going_down",
	entity.DetectStateBottom:    "bottom",
	entity.DetectStateGoingUp:   "going_up",
}

// FaceViewModel はフロント用の顔検出表示モデル。app 層で定義する。
type FaceViewModel struct {
	X            int     `json:"x"`
	Y            int     `json:"y"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	FrameWidth   int     `json:"frameWidth"`
	FrameHeight  int     `json:"frameHeight"`
	Ratio        float64 `json:"ratio"`
	State        string  `json:"state"`
	RepCompleted bool    `json:"repCompleted"`
}

// FaceViewModelFrom は usecase の出力（entity）を ViewModel に変換する。
func FaceViewModelFrom(out *usecase.WatchSquatOutput) *FaceViewModel {
	if out == nil || out.Face == nil || out.Judgement == nil {
		return nil
	}
	face := out.Face
	judgement := out.Judgement
	ratio := 0.0
	if face.FrameHeight > 0 {
		ratio = float64(face.CenterY()) / float64(face.FrameHeight)
	}
	return &FaceViewModel{
		X:            face.X,
		Y:            face.Y,
		Width:        face.Width,
		Height:       face.Height,
		FrameWidth:   face.FrameWidth,
		FrameHeight:  face.FrameHeight,
		Ratio:        ratio,
		State:        detectStateLabels[judgement.State],
		RepCompleted: judgement.IsRepCompleted,
	}
}

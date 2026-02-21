package entity

import "time"

type DetectState int

const (
	DetectStateUnknown DetectState = iota
	DetectStateStanding
	DetectStateGoingDown
	DetectStateBottom
	DetectStateGoingUp
)

const (
	DefaultTopRatio    = 0.7 // jusge going down ratio
	DefaultBottomRatio = 0.6 // judge going up ratio
)

type Judgement struct {
	Timestamp      time.Time
	State          DetectState
	IsRepCompleted bool
}

type Judgements []*Judgement

func NewJudgement(face *Face) *Judgement {
	return &Judgement{
		Timestamp: face.Timestamp,
		State:     DetectStateUnknown,
	}
}

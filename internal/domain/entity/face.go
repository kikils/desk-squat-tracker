package entity

import "time"

type Face struct {
	Timestamp   time.Time
	X           int // 左上 X
	Y           int // 左上 Y
	Width       int // 幅
	Height      int // 高さ
	FrameHeight int // フレームの高さ
	FrameWidth  int // フレームの幅
}

type Faces []*Face

func (m *Face) CenterY() int {
	return m.Y + m.Height/2
}

// TopY は顔の上端の Y 座標（判定用に顔の上の真ん中を使う）
func (m *Face) TopY() int {
	return m.Y
}

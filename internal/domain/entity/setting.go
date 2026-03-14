package entity

type Setting struct {
	TopRatio    float64 // しゃがみ始め判定（顔がこの比率より下に来たら GoingDown/Bottom）
	BottomRatio float64 // 立ち上がり判定（顔がこの比率より上に来たら GoingUp/Standing）
}

// DefaultSetting はデフォルトの設定を返す。
func DefaultSetting() *Setting {
	return &Setting{
		TopRatio:    DefaultTopRatio,
		BottomRatio: DefaultBottomRatio,
	}
}

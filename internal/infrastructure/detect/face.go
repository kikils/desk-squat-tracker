package detect

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"time"

	pigo "github.com/esimov/pigo/core"
	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
	"github.com/kikils/desk-squat-tracker/internal/errors"
)

//go:embed cascade/facefinder
var embeddedFacefinder []byte

// DefaultCascadeParams: ShiftFactor/ScaleFactor をやや細かくして検出精度を優先（pigo 推奨付近）。
// 参考: https://github.com/esimov/pigo
var DefaultCascadeParams = CascadeParams{
	MinSize:      20,
	MaxSize:      1000,
	ShiftFactor:  0.08,  // 0.1 より細かいグリッドで探索（精度↑、やや重い）
	ScaleFactor:  1.08,  // 1.1 より細かいスケール段階（サイズのブレ軽減）
	IOUThreshold: 0.18,  // クラスタをやや厳しめに（重複検出の統合）
	MinScore:     2.0,   // Q がこれ未満の検出は捨てる。0 で無効
}

type CascadeParams struct {
	MinSize      int
	MaxSize      int
	ShiftFactor  float64
	ScaleFactor  float64
	IOUThreshold float64
	MinScore     float32 // 検出スコア Q の最小値。これ未満は採用しない
}

type FaceRepository struct {
	pigo   *pigo.Pigo
	params CascadeParams
}

func NewFaceRepository() (repository.FaceRepository, error) {
	p := pigo.NewPigo()
	unpacked, err := p.Unpack(embeddedFacefinder)
	if err != nil {
		return nil, err
	}
	return &FaceRepository{
		pigo:   unpacked,
		params: DefaultCascadeParams,
	}, nil
}

func (r *FaceRepository) Detect(frame []byte, t time.Time) (*entity.Face, error) {
	img, _, err := image.Decode(bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}
	return r.DetectFromImage(img, t)
}

func (r *FaceRepository) DetectFromImage(img image.Image, t time.Time) (*entity.Face, error) {
	bounds := img.Bounds()
	cols := bounds.Dx()
	rows := bounds.Dy()
	pixels := pigo.RgbToGrayscale(img)

	cp := pigo.CascadeParams{
		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
		MinSize:     r.params.MinSize,
		MaxSize:     r.params.MaxSize,
		ShiftFactor: r.params.ShiftFactor,
		ScaleFactor: r.params.ScaleFactor,
	}

	dets := r.pigo.RunCascade(cp, 0)
	dets = r.pigo.ClusterDetections(dets, r.params.IOUThreshold)

	if len(dets) == 0 {
		return nil, errors.ErrNotFound.Errorf("not found face")
	}

	// クラスタ群のうちスコア Q が最大の検出を採用（pigo の Detection.Q = 信頼度）
	best := 0
	for i := 1; i < len(dets); i++ {
		if dets[i].Q > dets[best].Q {
			best = i
		}
	}
	d := &dets[best]
	if d.Q < r.params.MinScore {
		return nil, errors.ErrNotFound.Errorf("no face above MinScore")
	}

	return &entity.Face{
		X:           d.Col - d.Scale/2,
		Y:           d.Row - d.Scale/2,
		Width:       d.Scale,
		Height:      d.Scale,
		Timestamp:   t,
		FrameHeight: rows,
		FrameWidth:  cols,
	}, nil
}

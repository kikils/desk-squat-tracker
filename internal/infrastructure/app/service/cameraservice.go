package service

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"github.com/kikils/desk-squat-tracker/internal/infrastructure/camera"
	"github.com/kikils/desk-squat-tracker/internal/usecase"
	"github.com/kikils/desk-squat-tracker/internal/utils"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	jpegQuality   = 75
	previewFPS    = 5
	previewPeriod = time.Second / previewFPS
)

// CameraDevice はフロントに渡すカメラ情報。
type CameraDevice struct {
	Index int    `json:"Index"`
	Name  string `json:"Name"`
}

type CameraService struct {
	InputPort usecase.WatchSquatInputPort
	OnResult  func(*usecase.WatchSquatOutput)
	OnPreview func(dataURL string)

	ctx context.Context

	mu            sync.Mutex
	captureCancel context.CancelFunc
	captureGen    int // incremented per StartCapture; used to avoid clearing a newer capture's cancel
}

func (s *CameraService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.ctx = ctx
	return nil
}

// ListCameras は利用可能なカメラ一覧を返す。
func (s *CameraService) ListCameras() ([]CameraDevice, error) {
	devs, err := camera.ListDevices()
	if err != nil || len(devs) == 0 {
		return []CameraDevice{{Index: 0, Name: "デフォルトカメラ"}}, nil
	}
	out := make([]CameraDevice, len(devs))
	for i := range devs {
		out[i] = CameraDevice{Index: devs[i].Index, Name: devs[i].Name}
	}
	return out, nil
}

// StartCapture は指定したデバイスインデックスでキャプチャを開始する。macOS では AVFoundation で選択。
func (s *CameraService) StartCapture(deviceIndex int) error {
	s.mu.Lock()
	if s.captureCancel != nil {
		s.mu.Unlock()
		return nil
	}
	ctx, cancel := context.WithCancel(s.ctx)
	s.captureGen++
	gen := s.captureGen
	s.captureCancel = cancel
	s.mu.Unlock()

	frames, err := camera.StartStream(ctx, deviceIndex)
	if err != nil {
		cancel()
		s.mu.Lock()
		s.captureCancel = nil
		s.mu.Unlock()
		return err
	}

	go func() {
		defer func() {
			s.mu.Lock()
			if s.captureGen == gen {
				s.captureCancel = nil
			}
			s.mu.Unlock()
		}()
		lastPreview := time.Time{}
		for {
			select {
			case <-ctx.Done():
				return
			case f, ok := <-frames:
				if !ok {
					return
				}
				jpegBytes, err := utils.EncodeJPEG(utils.Frame{Data: f.Data, Width: f.Width, Height: f.Height}, jpegQuality)
				if err != nil || len(jpegBytes) == 0 {
					continue
				}
				if s.OnPreview != nil && time.Since(lastPreview) >= previewPeriod {
					lastPreview = time.Now()
					dataURL := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpegBytes)
					s.OnPreview(dataURL)
				}
				out, err := s.InputPort.Execute(ctx, jpegBytes, time.Now())
				if err != nil {
					continue
				}
				if s.OnResult != nil && out != nil && out.Face != nil && out.Judgement != nil {
					s.OnResult(out)
				}
			}
		}
	}()
	return nil
}

func (s *CameraService) StopCapture() {
	s.mu.Lock()
	cancel := s.captureCancel
	s.captureCancel = nil
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

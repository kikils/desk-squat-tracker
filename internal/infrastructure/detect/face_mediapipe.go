package detect

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/kikils/desk-squat-tracker/internal/config"
	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
	"github.com/kikils/desk-squat-tracker/internal/errors"
	"golang.org/x/xerrors"
)

func StartFaceDetectServer(ctx context.Context) error {
	var cmd *exec.Cmd
	// FIX: ビルド時に実行できるようにする
	if config.Get().FaceDetectServer.Debug {
		cmd = exec.CommandContext(ctx, "uv", "run", config.Get().FaceDetectServer.ScriptName, "--serve", "--port", strconv.Itoa(config.Get().FaceDetectServer.ServerPort))
		cmd.Dir = config.Get().FaceDetectServer.ScriptBasePath
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return xerrors.Errorf("face_mediapipe: start: %w", err)
	}

	for i := 0; i < 10; i++ {
		resp, err := http.Get(config.Get().FaceDetectServer.ServerURL() + "/health")
		if err != nil {
			if i == 9 {
				return xerrors.Errorf("face_mediapipe: health check: %w", err)
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			if i == 9 {
				return xerrors.New("face_mediapipe: health check failed")
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	return nil
}

type MediaPipeFaceRepository struct {
	client *http.Client
}

func NewMediaPipeFaceRepository() (repository.FaceRepository, error) {
	return &MediaPipeFaceRepository{
		client: &http.Client{Timeout: config.Get().FaceDetectServer.Timeout},
	}, nil
}

type mediaPipeResult struct {
	X           int `json:"x"`
	Y           int `json:"y"`
	Width       int `json:"width"`
	Height      int `json:"height"`
	FrameWidth  int `json:"frame_width"`
	FrameHeight int `json:"frame_height"`
}

func (r *MediaPipeFaceRepository) Detect(ctx context.Context, frame []byte, t time.Time) (*entity.Face, error) {
	url := config.Get().FaceDetectServer.ServerURL() + "/detect"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(frame))
	if err != nil {
		return nil, xerrors.Errorf("face_mediapipe: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, xerrors.Errorf("face_mediapipe: request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrNotFound.Errorf("face_mediapipe: server returned %d", resp.StatusCode)
	}
	var res mediaPipeResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, xerrors.Errorf("face_mediapipe: decode json: %w", err)
	}
	if res.FrameWidth == 0 || res.FrameHeight == 0 {
		return nil, errors.ErrNotFound.Errorf("face_mediapipe: no face")
	}
	return &entity.Face{
		X:           res.X,
		Y:           res.Y,
		Width:       res.Width,
		Height:      res.Height,
		FrameWidth:  res.FrameWidth,
		FrameHeight: res.FrameHeight,
		Timestamp:   t,
	}, nil
}

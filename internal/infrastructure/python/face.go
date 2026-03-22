package python

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/kikils/desk-squat-tracker/internal/config"
	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	"github.com/kikils/desk-squat-tracker/internal/domain/repository"
	"github.com/kikils/desk-squat-tracker/internal/errors"
	"golang.org/x/xerrors"
)

const healthRetries = 10
const healthRetryDelay = 3 * time.Second

var (
	faceChildMu sync.Mutex
	faceChild   *exec.Cmd
)

func StopFaceDetectServer() {
	faceChildMu.Lock()
	cmd := faceChild
	faceChild = nil
	faceChildMu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return
	}
	faceChildKill(cmd)
	_ = cmd.Wait()
}

func StartFaceDetectServer(ctx context.Context) (err error) {
	cfg := config.Get().FaceDetectServer
	if !cfg.Debug {
		return waitFaceDetectHealth(cfg.ServerURL() + "/health")
	}

	cmd, err := buildFaceDetectCmd(ctx, cfg)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return xerrors.Errorf("face_mediapipe: start: %w", err)
	}
	faceChildMu.Lock()
	faceChild = cmd
	faceChildMu.Unlock()
	defer func() {
		if err != nil {
			StopFaceDetectServer()
		}
	}()

	err = waitFaceDetectHealth(cfg.ServerURL() + "/health")
	return err
}

func buildFaceDetectCmd(ctx context.Context, cfg config.FaceDetectServer) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	if p := resolveBundledFaceDetect(); p != "" {
		cmd = exec.CommandContext(ctx, p, "--serve", "--port", strconv.Itoa(cfg.ServerPort))
		cmd.Dir = filepath.Dir(p)
	} else {
		helperDir, err := resolveHelperDirNextToBin()
		if err != nil {
			return nil, xerrors.Errorf("face_mediapipe: helper dir: %w", err)
		}
		cmd = exec.CommandContext(ctx, "uv", "run", cfg.ScriptName, "--serve", "--port", strconv.Itoa(cfg.ServerPort))
		cmd.Dir = helperDir
	}
	faceChildPrepare(cmd)
	return cmd, nil
}

func waitFaceDetectHealth(healthURL string) error {
	for i := 0; i < healthRetries; i++ {
		resp, err := http.Get(healthURL)
		if err != nil {
			if i == healthRetries-1 {
				return xerrors.Errorf("face_mediapipe: health check: %w", err)
			}
			time.Sleep(healthRetryDelay)
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			if i == healthRetries-1 {
				return xerrors.New("face_mediapipe: health check failed")
			}
			time.Sleep(healthRetryDelay)
			continue
		}
		return nil
	}
	return xerrors.New("face_mediapipe: health check failed")
}

func resolveBundledFaceDetect() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	p := filepath.Join(filepath.Dir(exe), "face_detect")
	if _, err := os.Stat(p); err != nil {
		return ""
	}
	return p
}

func resolveHelperDirNextToBin() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	helper := filepath.Join(filepath.Dir(exe), "..", "helper")
	abs, err := filepath.Abs(helper)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(abs); err != nil {
		return "", err
	}
	return abs, nil
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

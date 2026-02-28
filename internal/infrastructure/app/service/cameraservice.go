package service

import (
	"context"
	"encoding/base64"
	_ "image/jpeg"
	"strings"
	"time"

	"github.com/kikils/desk-squat-tracker/internal/usecase"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type CameraService struct {
	InputPort usecase.WatchSquatInputPort
	OnResult  func(*usecase.WatchSquatOutput)

	ctx context.Context
}

func (c *CameraService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	c.ctx = ctx
	return nil
}

func (c *CameraService) ReceiveFrame(frameBase64 string) error {
	raw := frameBase64
	if idx := strings.Index(frameBase64, ","); idx >= 0 {
		raw = frameBase64[idx+1:]
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return err
	}

	out, err := c.InputPort.Execute(c.ctx, decoded, time.Now())
	if err != nil {
		return err
	}
	if c.OnResult != nil && out != nil && out.Face != nil && out.Judgement != nil {
		c.OnResult(out)
	}
	return nil
}

package camera

import (
	"context"
	"fmt"
	"runtime"
)

// Frame は 1 フレーム（YCbCr444 packed、Width*Height*3 バイト）。
type Frame struct {
	Data   []byte
	Width  int
	Height int
}

// StartStream は指定デバイスでキャプチャを開始し、フレームチャネルを返す。macOS のみ対応。
func StartStream(ctx context.Context, deviceIndex int) (<-chan Frame, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("camera: unsupported platform %s", runtime.GOOS)
	}
	devs, _ := ListDevices()
	if deviceIndex < 0 || deviceIndex >= len(devs) {
		return nil, fmt.Errorf("camera: device index %d out of range (0..%d)", deviceIndex, len(devs)-1)
	}
	return startStreamDarwin(ctx, deviceIndex)
}

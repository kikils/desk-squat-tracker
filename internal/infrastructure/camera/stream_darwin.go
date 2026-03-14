//go:build darwin

package camera

import "context"

// startStreamDarwin は macOS で指定インデックスのカメラからストリームを開始する。
// getVideoDevices() と同一の DiscoverySession により、一覧の index と対応する。
func startStreamDarwin(ctx context.Context, deviceIndex int) (<-chan Frame, error) {
	return startStreamDarwinCGO(ctx, deviceIndex)
}

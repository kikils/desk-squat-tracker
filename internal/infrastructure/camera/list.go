package camera

import (
	"runtime"
)

// Device は利用可能なカメラデバイスを表す。
type Device struct {
	Index int    // 0-based index（StartCapture に渡す値）
	Name  string // 表示名
	ID    string // プラットフォーム固有の ID（macOS では未使用）
}

// ListDevices は利用可能なカメラデバイス一覧を返す。macOS のみ AVFoundation で一覧取得。
func ListDevices() ([]Device, error) {
	return listDevices()
}

func listDevices() ([]Device, error) {
	if runtime.GOOS == "darwin" {
		return listDevicesDarwin()
	}
	return listDevicesDefault()
}

func listDevicesDefault() ([]Device, error) {
	return []Device{{Index: 0, Name: "デフォルトカメラ", ID: ""}}, nil
}

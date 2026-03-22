package service

import "github.com/wailsapp/wails/v3/pkg/application"

type AppService struct{}

// Quit はアプリ全体を終了します（トレイ常駐も含めてプロセスが終了します）。
// バインディングの応答が WebView に返る前に同期で Quit するとデッドロックするため、
// メインスレッドへ非同期で渡してから終了処理を走らせます。
func (s *AppService) Quit() {
	application.InvokeAsync(func() {
		application.Get().Quit()
	})
}

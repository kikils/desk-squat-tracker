package app

import (
	"io/fs"
	"time"

	"github.com/kikils/desk-squat-tracker/internal/config"
	"github.com/kikils/desk-squat-tracker/internal/domain/entity"
	dservice "github.com/kikils/desk-squat-tracker/internal/domain/service"
	"github.com/kikils/desk-squat-tracker/internal/infrastructure/app/service"
	"github.com/kikils/desk-squat-tracker/internal/infrastructure/file"
	"github.com/kikils/desk-squat-tracker/internal/infrastructure/memory"
	"github.com/kikils/desk-squat-tracker/internal/infrastructure/python"
	"github.com/kikils/desk-squat-tracker/internal/usecase"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

func init() {
	application.RegisterEvent[string]("time")
	application.RegisterEvent[int]("squat")
	application.RegisterEvent[*FaceViewModel]("face")
	application.RegisterEvent[string]("cameraPreview")
}

func Run(assets fs.FS, iconStandup, iconSquat []byte) error {
	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.

	faceRepository, err := python.NewMediaPipeFaceRepository()
	if err != nil {
		return err
	}

	judgementRepository := memory.NewJudgementRepository()
	settingRepository, err := file.NewSettingRepository()
	if err != nil {
		return err
	}
	squatJudger := dservice.NewSquatJudger(faceRepository, judgementRepository, settingRepository)
	cameraSvc := &service.CameraService{
		InputPort: usecase.NewWatchSquatUsecase(faceRepository, judgementRepository, squatJudger),
	}
	statsSvc := &service.StatsService{
		InputPort: usecase.NewGetStatsUsecase(judgementRepository),
	}
	settingsSvc := &service.SettingsService{
		GetSettingInputPort:    usecase.NewGetSettingUsecase(settingRepository),
		UpdateSettingInputPort: usecase.NewUpdateSettingUsecase(settingRepository),
	}

	app := application.New(application.Options{
		Name:        config.AppName,
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&service.GreetService{}),
			application.NewService(cameraSvc),
			application.NewService(statsSvc),
			application.NewService(settingsSvc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			// トレイに格納した状態でクリックしても終了しないよう false にする。
			// ウィンドウを閉じたときは Hide にしているため、アプリはトレイで動作し続ける。
			ApplicationShouldTerminateAfterLastWindowClosed: false,
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	})

	systray := app.SystemTray.New()
	systray.SetIcon(iconStandup)

	cameraSvc.OnResult = func(out *usecase.WatchSquatOutput) {
		if vm := FaceViewModelFrom(out); vm != nil {
			app.Event.Emit("face", vm)
		}
		if out != nil && out.Judgement != nil {
			if out.Judgement.IsRepCompleted {
				app.Event.Emit("squat", 0)
			}
			// 判定状態に応じてトレイアイコンを切り替え（立っている→standup、しゃがんでいる→squat）
			if out.Judgement.State == entity.DetectStateStanding {
				systray.SetIcon(iconStandup)
			} else {
				systray.SetIcon(iconSquat)
			}
		}
	}
	cameraSvc.OnPreview = func(dataURL string) {
		app.Event.Emit("cameraPreview", dataURL)
	}

	popupWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Width:           400,
		Height:          400,
		Name:            "Popup",
		Frameless:       true,
		AlwaysOnTop:     true,
		Hidden:          true,
		DisableResize:   true,
		HideOnFocusLost: true, // 他のウィンドウをクリックしたら非表示にする
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	popupWindow.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		popupWindow.Hide()
		event.Cancel()
	})

	systray.AttachWindow(popupWindow).WindowOffset(2)

	// Create a goroutine that emits an event containing the current time every second.
	// The frontend can listen to this event and update the UI accordingly.
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	if err := python.StartFaceDetectServer(app.Context()); err != nil {
		return err
	}

	// Run the application. This blocks until the application has been exited.
	return app.Run()
}

package app

import (
	"io/fs"
	"time"

	dservice "github.com/kikils/desk-squat-tracker/internal/domain/service"
	"github.com/kikils/desk-squat-tracker/internal/infrastructure/app/service"
	"github.com/kikils/desk-squat-tracker/internal/infrastructure/detect"
	"github.com/kikils/desk-squat-tracker/internal/infrastructure/memory"
	"github.com/kikils/desk-squat-tracker/internal/usecase"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func init() {
	application.RegisterEvent[string]("time")
	application.RegisterEvent[int]("squat")
	application.RegisterEvent[*FaceViewModel]("face")
}

func Run(assets fs.FS, iconBytes []byte) error {
	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.

	faceRepository, err := detect.NewMediaPipeFaceRepository()
	if err != nil {
		return err
	}

	judgementRepository := memory.NewJudgementRepository()
	squatJudger := dservice.NewSquatJudger(faceRepository, judgementRepository)
	cameraSvc := &service.CameraService{
		InputPort: usecase.NewWatchSquatUsecase(faceRepository, judgementRepository, squatJudger),
	}

	app := application.New(application.Options{
		Name:        "desk-squat-tracker",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&service.GreetService{}),
			application.NewService(cameraSvc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	cameraSvc.OnResult = func(out *usecase.WatchSquatOutput) {
		if vm := FaceViewModelFrom(out); vm != nil {
			app.Event.Emit("face", vm)
		}
	}

	systray := app.SystemTray.New()
	systray.SetIcon(iconBytes)

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Window 1",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// Create a goroutine that emits an event containing the current time every second.
	// The frontend can listen to this event and update the UI accordingly.
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	if err := detect.StartFaceDetectServer(app.Context()); err != nil {
		return err
	}

	// Run the application. This blocks until the application has been exited.
	return app.Run()
}

package config

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var (
	conf Config
	once sync.Once
)

type Config struct {
	FaceDetectServer FaceDetectServer
}

type FaceDetectServer struct {
	ServerHost     string        `default:"127.0.0.1"`
	ServerPort     int           `default:"8765"`
	ScriptBasePath string        `default:"helper"`
	ScriptName     string        `default:"face_detect_mediapipe.py"`
	Timeout        time.Duration `default:"10s"`
	Debug          bool          `default:"true"`
}

func (f FaceDetectServer) ServerURL() string {
	return fmt.Sprintf("http://%s:%d", f.ServerHost, f.ServerPort)
}

func Get() Config {
	once.Do(func() {
		if err := envconfig.Process("facedetectserver", &conf.FaceDetectServer); err != nil {
			log.Fatal(err.Error())
		}
	})
	return conf
}

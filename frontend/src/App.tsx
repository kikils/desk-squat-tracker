import { useState, useEffect } from 'react'
import {Events, WML} from "@wailsio/runtime";
import {GreetService} from "../bindings/github.com/kikils/desk-squat-tracker/internal/infrastructure/app/service";
import { useCameraStream } from "./hooks/useCameraStream";

export interface FaceDetectedPayload {
  x: number;
  y: number;
  width: number;
  height: number;
  frameWidth: number;
  frameHeight: number;
  ratio: number;
  state: string;
  repCompleted: boolean;
}

function App() {
  const [name, setName] = useState<string>('');
  const [result, setResult] = useState<string>('Please enter your name below 👇');
  const [time, setTime] = useState<string>('Listening for Time event...');
  const [faceData, setFaceData] = useState<FaceDetectedPayload | null>(null);

  const { videoRef, canvasRef, isActive, error, start, stop } = useCameraStream();

  const doGreet = () => {
    let localName = name;
    if (!localName) {
      localName = 'anonymous';
    }
    GreetService.Greet(localName).then((resultValue: string) => {
      setResult(resultValue);
    }).catch((err: any) => {
      console.log(err);
    });
  }

  useEffect(() => {
    Events.On('time', (timeValue: any) => {
      setTime(timeValue.data);
    });
    Events.On('face', (ev: any) => {
      const payload = ev.data as FaceDetectedPayload;
      if (payload && typeof payload.ratio === 'number') {
        setFaceData(payload);
      }
    });
    // Reload WML so it picks up the wml tags
    WML.Reload();
  }, []);

  return (
    <div className="container">
      <div>
        <a data-wml-openURL="https://wails.io">
          <img src="/wails.png" className="logo" alt="Wails logo"/>
        </a>
        <a data-wml-openURL="https://reactjs.org">
          <img src="/react.svg" className="logo react" alt="React logo"/>
        </a>
      </div>
      <h1>Wails + React</h1>
      <div className="result">{result}</div>
      <div className="card">
        <div className="input-box">
          <input className="input" value={name} onChange={(e) => setName(e.target.value)} type="text" autoComplete="off"/>
          <button className="btn" onClick={doGreet}>Greet</button>
        </div>
      </div>

      <div className="card">
        <h2>カメラ</h2>
        <canvas ref={canvasRef} style={{ display: "none" }} />
        <div style={{ position: "relative", display: "inline-block" }}>
          <video ref={videoRef} muted playsInline style={{ width: 320, maxWidth: "100%", background: "#1b2636", borderRadius: 8, display: "block" }} />
          {faceData && faceData.frameWidth > 0 && faceData.frameHeight > 0 && (
            <div
              style={{
                position: "absolute",
                top: 0,
                left: 0,
                width: "100%",
                height: "100%",
                pointerEvents: "none",
                boxSizing: "border-box",
              }}
            >
              <div
                style={{
                  position: "absolute",
                  left: `${(faceData.x / faceData.frameWidth) * 100}%`,
                  top: `${(faceData.y / faceData.frameHeight) * 100}%`,
                  width: `${(faceData.width / faceData.frameWidth) * 100}%`,
                  height: `${(faceData.height / faceData.frameHeight) * 100}%`,
                  border: "2px solid #2ecc71",
                  borderRadius: 4,
                  boxSizing: "border-box",
                }}
              />
            </div>
          )}
        </div>
        {error && <p style={{ color: "#e74c3c" }}>{error}</p>}
        <div style={{ marginTop: 8 }}>
          {!isActive ? (
            <button className="btn" onClick={start}>カメラ開始</button>
          ) : (
            <button className="btn" onClick={stop}>カメラ停止</button>
          )}
        </div>
        {faceData && (
          <div style={{ marginTop: 12, padding: 12, background: "#1b2636", borderRadius: 8, fontSize: 14 }}>
            <h3 style={{ marginTop: 0 }}>顔検出</h3>
            <p style={{ margin: "4px 0" }}>座標: x={faceData.x}, y={faceData.y}</p>
            <p style={{ margin: "4px 0" }}>サイズ: {faceData.width} × {faceData.height}</p>
            <p style={{ margin: "4px 0" }}>フレーム: {faceData.frameWidth} × {faceData.frameHeight}</p>
            <p style={{ margin: "4px 0" }}>ratio: {faceData.ratio.toFixed(3)}</p>
            <p style={{ margin: "4px 0" }}>状態: {faceData.state}</p>
            {faceData.repCompleted && <p style={{ margin: "4px 0", color: "#2ecc71" }}>✓ 1 rep 完了</p>}
          </div>
        )}
      </div>
      <div className="footer">
        <div><p>Click on the Wails logo to learn more</p></div>
        <div><p>{time}</p></div>
      </div>
    </div>
  )
}

export default App

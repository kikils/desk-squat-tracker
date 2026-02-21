import { useCallback, useRef, useState } from "react";
import { CameraService } from "../../bindings/github.com/kikils/desk-squat-tracker/internal/infrastructure/app/service";

const CAPTURE_INTERVAL_MS = 100; // 10 fps で Go に送信
const VIDEO_WIDTH = 640;
const VIDEO_HEIGHT = 480;
const JPEG_QUALITY = 0.7;

export function useCameraStream() {
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const [isActive, setIsActive] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const captureAndSend = useCallback(() => {
    const video = videoRef.current;
    const canvas = canvasRef.current;
    if (!video || !canvas || video.readyState !== video.HAVE_ENOUGH_DATA) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    ctx.drawImage(video, 0, 0);
    const dataUrl = canvas.toDataURL("image/jpeg", JPEG_QUALITY);
    CameraService.ReceiveFrame(dataUrl).catch((err) => {
      console.warn("ReceiveFrame error:", err);
    });
  }, []);

  const start = useCallback(async () => {
    setError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { width: VIDEO_WIDTH, height: VIDEO_HEIGHT },
      });
      streamRef.current = stream;
      const video = videoRef.current;
      if (video) {
        video.srcObject = stream;
        await video.play();
      }
      intervalRef.current = setInterval(captureAndSend, CAPTURE_INTERVAL_MS);
      setIsActive(true);
    } catch (e) {
      const msg = e instanceof Error ? e.message : "カメラを起動できませんでした";
      setError(msg);
      setIsActive(false);
    }
  }, [captureAndSend]);

  const stop = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    const stream = streamRef.current;
    if (stream) {
      stream.getTracks().forEach((t) => t.stop());
      streamRef.current = null;
    }
    const video = videoRef.current;
    if (video) {
      video.srcObject = null;
    }
    setIsActive(false);
  }, []);

  return {
    videoRef,
    canvasRef,
    isActive,
    error,
    start,
    stop,
  };
}

import { useCallback, useState } from "react";
import { CameraService } from "../../bindings/github.com/kikils/desk-squat-tracker/internal/infrastructure/app/service";

export function useCameraStream() {
  const [isActive, setIsActive] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const start = useCallback(async (deviceIndex: number = 0) => {
    setError(null);
    try {
      await CameraService.StartCapture(deviceIndex);
      setIsActive(true);
    } catch (e) {
      const msg = e instanceof Error ? e.message : "カメラを起動できませんでした";
      setError(msg);
      setIsActive(false);
    }
  }, []);

  const stop = useCallback(() => {
    CameraService.StopCapture();
    setIsActive(false);
  }, []);

  return {
    isActive,
    error,
    start,
    stop,
  };
}

import { useState, useEffect, useCallback, useRef } from 'react'
import { Events, WML } from "@wailsio/runtime";
import { AppService, CameraService, SettingsService, StatsService, type CameraDevice } from "../bindings/github.com/kikils/desk-squat-tracker/internal/infrastructure/app/service";
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

type Page = 'summary' | 'camera';

const PAGE_LABELS: Record<Page, string> = {
  summary: '成績',
  camera: 'カメラ設定',
};

function App() {
  const [page, setPage] = useState<Page>('summary');
  const [todayCount, setTodayCount] = useState<number | null>(null);
  const [faceData, setFaceData] = useState<FaceDetectedPayload | null>(null);
  const [previewDataUrl, setPreviewDataUrl] = useState<string | null>(null);
  const [topRatio, setTopRatio] = useState<number>(0.7);
  const [bottomRatio, setBottomRatio] = useState<number>(0.6);
  const [settingLoaded, setSettingLoaded] = useState(false);
  const [draggingLine, setDraggingLine] = useState<'top' | 'bottom' | null>(null);
  const [cameras, setCameras] = useState<CameraDevice[]>([]);
  const [selectedCameraIndex, setSelectedCameraIndex] = useState(0);
  const [quitConfirmOpen, setQuitConfirmOpen] = useState(false);
  const overlayRef = useRef<HTMLDivElement>(null);
  const ratiosRef = useRef({ topRatio: 0.7, bottomRatio: 0.6 });
  const lastRatiosRef = useRef({ topRatio: 0.7, bottomRatio: 0.6 });
  const userHasChangedRef = useRef(false);
  ratiosRef.current = { topRatio, bottomRatio };
  if (draggingLine === null) lastRatiosRef.current = { topRatio, bottomRatio };

  const { isActive, error, start, stop } = useCameraStream();

  const clampRatio = useCallback((value: number) => Math.max(0, Math.min(1, value)), []);

  const getRatioFromEvent = useCallback((clientY: number): number => {
    const el = overlayRef.current;
    if (!el) return 0;
    const rect = el.getBoundingClientRect();
    return clampRatio((clientY - rect.top) / rect.height);
  }, [clampRatio]);

  const handleLinePointerDown = useCallback((line: 'top' | 'bottom', e: React.PointerEvent) => {
    e.preventDefault();
    userHasChangedRef.current = true;
    (e.target as HTMLElement).setPointerCapture?.(e.pointerId);
    setDraggingLine(line);
  }, []);

  const stepRatio = 0.02;
  const handleLineKeyDown = useCallback((line: 'top' | 'bottom', e: React.KeyboardEvent) => {
    if (e.key !== 'ArrowUp' && e.key !== 'ArrowDown') return;
    e.preventDefault();
    userHasChangedRef.current = true;
    const sign = e.key === 'ArrowDown' ? 1 : -1;
    if (line === 'top') {
      setTopRatio((prev) => {
        const next = clampRatio(prev + sign * stepRatio);
        return next > bottomRatio + 0.01 ? next : prev;
      });
    } else {
      setBottomRatio((prev) => {
        const next = clampRatio(prev + sign * stepRatio);
        return next < topRatio - 0.01 ? next : prev;
      });
    }
  }, [clampRatio, topRatio, bottomRatio]);

  const handlePointerMove = useCallback((e: PointerEvent) => {
    if (draggingLine === null) return;
    const ratio = getRatioFromEvent(e.clientY);
    const { topRatio: t, bottomRatio: b } = ratiosRef.current;
    if (draggingLine === 'top') {
      const v = ratio > b + 0.01 ? ratio : b + 0.01;
      setTopRatio(v);
      lastRatiosRef.current = { ...lastRatiosRef.current, topRatio: v };
    } else {
      const v = ratio < t - 0.01 ? ratio : t - 0.01;
      setBottomRatio(v);
      lastRatiosRef.current = { ...lastRatiosRef.current, bottomRatio: v };
    }
  }, [draggingLine, getRatioFromEvent]);

  const handlePointerUp = useCallback(() => {
    if (draggingLine === null) return;
    setDraggingLine(null);
  }, [draggingLine]);

  const AUTO_SAVE_DELAY_MS = 400;
  useEffect(() => {
    if (!settingLoaded || !userHasChangedRef.current || draggingLine !== null) return;
    const id = setTimeout(() => {
      const { topRatio: t, bottomRatio: b } = lastRatiosRef.current;
      SettingsService.UpdateSetting(t, b).catch((err) =>
        console.warn('UpdateSetting error:', err)
      );
    }, AUTO_SAVE_DELAY_MS);
    return () => clearTimeout(id);
  }, [settingLoaded, topRatio, bottomRatio, draggingLine]);

  useEffect(() => {
    if (draggingLine === null) return;
    window.addEventListener('pointermove', handlePointerMove);
    window.addEventListener('pointerup', handlePointerUp);
    window.addEventListener('pointercancel', handlePointerUp);
    return () => {
      window.removeEventListener('pointermove', handlePointerMove);
      window.removeEventListener('pointerup', handlePointerUp);
      window.removeEventListener('pointercancel', handlePointerUp);
    };
  }, [draggingLine, handlePointerMove, handlePointerUp]);

  const fetchTodayStats = useCallback(() => {
    StatsService.GetStats(new Date().toISOString())
      .then((out) => {
        if (out) setTodayCount(out.RepCount);
      })
      .catch((err) => console.warn('GetStats error:', err));
  }, []);

  useEffect(() => {
    if (page === 'summary') {
      fetchTodayStats();
    }
  }, [page, fetchTodayStats]);

  const fetchSetting = useCallback(() => {
    SettingsService.GetSetting()
      .then((out) => {
        if (out) {
          setTopRatio(out.TopRatio);
          setBottomRatio(out.BottomRatio);
        }
        setSettingLoaded(true);
      })
      .catch((err) => {
        console.warn('GetSetting error:', err);
        setSettingLoaded(true);
      });
  }, []);

  useEffect(() => {
    if (page === 'camera' && !settingLoaded) {
      fetchSetting();
    }
  }, [page, settingLoaded, fetchSetting]);

  useEffect(() => {
    if (page === 'camera') {
      CameraService.ListCameras()
        .then((list) => {
          const devs = list ?? [];
          setCameras(devs);
          if (devs.length > 0) {
            setSelectedCameraIndex((prev) => (prev < devs.length ? prev : 0));
          }
        })
        .catch(() => setCameras([]));
    }
  }, [page]);

  useEffect(() => {
    Events.On('face', (ev: { data?: FaceDetectedPayload | null }) => {
      const payload = ev.data;
      if (payload && typeof payload.ratio === 'number') {
        setFaceData(payload as FaceDetectedPayload);
      }
    });
    Events.On('squat', () => {
      fetchTodayStats();
    });
    Events.On('cameraPreview', (ev: { data?: string }) => {
      if (typeof ev.data === 'string') setPreviewDataUrl(ev.data);
    });
    WML.Reload();
  }, [fetchTodayStats]);

  useEffect(() => {
    if (!isActive) setPreviewDataUrl(null);
  }, [isActive]);

  useEffect(() => {
    if (page === 'camera') {
      document.body.classList.add('camera-page-active');
      return () => document.body.classList.remove('camera-page-active');
    }
  }, [page]);

  const handleNavKey = (e: React.KeyboardEvent, targetPage: Page) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      setPage(targetPage);
    }
  };

  const handleCameraSwitch = () => {
    if (isActive) stop();
    else start(selectedCameraIndex);
  };

  const handleSwitchKey = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleCameraSwitch();
    }
  };

  useEffect(() => {
    if (!quitConfirmOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setQuitConfirmOpen(false);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [quitConfirmOpen]);

  const handleQuitConfirm = () => {
    AppService.Quit().catch((err) => console.warn('Quit error:', err));
  };

  return (
    <>
      <a href="#main" className="skip-link">
        メインコンテンツへ
      </a>
      <div className={`app-container${page === 'camera' ? ' app-container--camera-page' : ''}`} role="presentation">
        <header className="app-header">
          <span className="app-header-label">カメラを有効化</span>
          <button
            type="button"
            role="switch"
            aria-checked={isActive}
            aria-label={isActive ? 'カメラをオフにする' : 'カメラをオンにする'}
            className={`camera-switch ${isActive ? 'camera-switch--on' : ''}`}
            onClick={handleCameraSwitch}
            onKeyDown={handleSwitchKey}
          >
            <span className="camera-switch__thumb" aria-hidden />
          </button>
        </header>

        <nav className="app-nav" aria-label="ページ切り替え">
          {(['summary', 'camera'] as const).map((p) => (
            <button
              key={p}
              type="button"
              className="app-nav-btn"
              aria-current={page === p ? 'page' : undefined}
              aria-label={PAGE_LABELS[p]}
              onClick={() => setPage(p)}
              onKeyDown={(e) => handleNavKey(e, p)}
            >
              {PAGE_LABELS[p]}
            </button>
          ))}
        </nav>

        <main id="main" tabIndex={-1}>
          {page === 'summary' && (
            <section
              className="card card--summary reveal-summary"
              aria-labelledby="summary-heading"
              aria-busy={todayCount === null}
            >
              <h2 id="summary-heading">今日のスクワット</h2>
              <p style={{ margin: 0 }}>
                <span
                  className="stats-count"
                  aria-busy={todayCount === null}
                  aria-live="polite"
                  aria-label={todayCount !== null ? `今日の回数は${todayCount}回` : '読み込み中'}
                >
                  {todayCount !== null ? todayCount : '—'}
                </span>
                <span className="stats-unit">回</span>
              </p>
            </section>
          )}

          {page === 'camera' && (
            <section
              className="page-section page-section--camera"
              aria-labelledby="camera-heading"
              aria-describedby={error ? 'camera-error' : undefined}
            >
              <h2 id="camera-heading">カメラ設定</h2>
              <div className="camera-select-row">
                <label htmlFor="camera-select" className="camera-select-label">
                  使用するカメラ
                </label>
                <select
                  id="camera-select"
                  className="camera-select"
                  value={selectedCameraIndex}
                  onChange={(e) => setSelectedCameraIndex(Number(e.target.value))}
                  disabled={isActive}
                  aria-label="使用するカメラを選択"
                >
                  {cameras.length === 0 ? (
                    <option value={0}>読み込み中…</option>
                  ) : (
                    cameras.map((cam) => (
                      <option key={cam.Index} value={cam.Index}>
                        {cam.Name}
                      </option>
                    ))
                  )}
                </select>
              </div>
              <div className="camera-wrap">
                <div className="camera-viewport">
                  {!isActive && (
                    <div className="camera-placeholder" aria-hidden="true">
                      カメラ未開始
                    </div>
                  )}
                  {isActive && !previewDataUrl && (
                    <div className="camera-placeholder" aria-live="polite" aria-busy="true">
                      カメラ稼働中…
                    </div>
                  )}
                  {isActive && previewDataUrl && (
                    <img
                      src={previewDataUrl}
                      alt="カメラプレビュー"
                      className="camera-preview-img"
                      width={320}
                      height={240}
                      decoding="async"
                    />
                  )}
                </div>
                <div
                  ref={overlayRef}
                  className="ratio-lines-overlay"
                  aria-hidden={!settingLoaded}
                >
                  <div
                    className={`ratio-line ratio-line--top ${draggingLine === 'top' ? 'ratio-line--dragging' : ''}`}
                    style={{ top: `${topRatio * 100}%` }}
                    role="slider"
                    aria-label="しゃがみ始め判定の線（ドラッグで位置変更）"
                    aria-valuenow={Math.round(topRatio * 100)}
                    aria-valuemin={0}
                    aria-valuemax={100}
                    tabIndex={settingLoaded ? 0 : -1}
                    onPointerDown={(e) => handleLinePointerDown('top', e)}
                    onKeyDown={(e) => handleLineKeyDown('top', e)}
                  >
                    <span className="ratio-line__label">しゃがみ始め判定</span>
                  </div>
                  <div
                    className={`ratio-line ratio-line--bottom ${draggingLine === 'bottom' ? 'ratio-line--dragging' : ''}`}
                    style={{ top: `${bottomRatio * 100}%` }}
                    role="slider"
                    aria-label="立ち上がり判定の線（ドラッグで位置変更）"
                    aria-valuenow={Math.round(bottomRatio * 100)}
                    aria-valuemin={0}
                    aria-valuemax={100}
                    tabIndex={settingLoaded ? 0 : -1}
                    onPointerDown={(e) => handleLinePointerDown('bottom', e)}
                    onKeyDown={(e) => handleLineKeyDown('bottom', e)}
                  >
                    <span className="ratio-line__label">立ち上がり判定</span>
                  </div>
                </div>
                {faceData && faceData.frameWidth > 0 && faceData.frameHeight > 0 && (
                  <div className="face-overlay" aria-hidden="true">
                    <div
                      className="face-box"
                      style={{
                        left: `${(faceData.x / faceData.frameWidth) * 100}%`,
                        top: `${(faceData.y / faceData.frameHeight) * 100}%`,
                        width: `${(faceData.width / faceData.frameWidth) * 100}%`,
                        height: `${(faceData.height / faceData.frameHeight) * 100}%`,
                      }}
                    />
                  </div>
                )}
                {faceData && (
                  <div className="face-status-inline" aria-live="polite">
                    <span className="face-status-inline__state">状態: {faceData.state}</span>
                    {faceData.repCompleted && (
                      <span className="face-status-inline__rep rep-done">✓ 1 rep 完了</span>
                    )}
                  </div>
                )}
              </div>
              {error && (
                <p id="camera-error" className="error-msg" role="alert">
                  {error}
                </p>
              )}
            </section>
          )}
        </main>

        <footer className="app-footer">
          {!quitConfirmOpen ? (
            <button
              type="button"
              className="app-quit-btn app-quit-btn--exit"
              onClick={() => setQuitConfirmOpen(true)}
              aria-label="アプリを終了する"
            >
              <span className="app-quit-btn__row">
                <svg
                  className="app-quit-btn__icon"
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  aria-hidden
                >
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                  <polyline points="16 17 21 12 16 7" />
                  <line x1="21" y1="12" x2="9" y2="12" />
                </svg>
                <span className="app-quit-btn__label">アプリを終了</span>
              </span>
            </button>
          ) : (
            <div className="app-quit-confirm" role="group" aria-label="アプリ終了の確認">
              <p className="app-quit-confirm__msg" id="quit-confirm-desc">
                終了しますか？メニューバー／トレイからも消えます。
              </p>
              <div className="app-quit-confirm__actions" aria-describedby="quit-confirm-desc">
                <button
                  type="button"
                  className="app-quit-btn app-quit-btn--secondary"
                  onClick={() => setQuitConfirmOpen(false)}
                >
                  キャンセル
                </button>
                <button
                  type="button"
                  className="app-quit-btn app-quit-btn--danger"
                  onClick={handleQuitConfirm}
                >
                  終了する
                </button>
              </div>
            </div>
          )}
        </footer>
      </div>
    </>
  );
}

export default App;

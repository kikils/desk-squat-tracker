#!/usr/bin/env python3
"""
Face detection using MediaPipe.
- One-shot: python face_detect_mediapipe.py [image_path]  # or stdin
  Outputs one JSON line to stdout.
- Server (RPC): python face_detect_mediapipe.py --serve [--port PORT]
  HTTP server: POST /detect with body=image bytes, response=JSON.
  GET /health for readiness. Model loaded once at startup.
"""
import argparse
import json
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from threading import Lock

import cv2
import numpy as np
import mediapipe as mp

# Optional: allow running without mediapipe for health check
try:
    mp_face_detection = mp.solutions.face_detection
except Exception:
    mp_face_detection = None


def detect_one(image: "np.ndarray", face_detection: "mp.solutions.face_detection.FaceDetection | None" = None) -> dict | None:
    """Run face detection on one image. Returns dict or None if no face.
    If face_detection is provided (e.g. from server), it is reused; else a new one is created (one-shot)."""
    if image is None or image.size == 0:
        return None
    h, w = image.shape[:2]
    rgb = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
    if face_detection is not None:
        results = face_detection.process(rgb)
    else:
        with mp_face_detection.FaceDetection(
            model_selection=0, min_detection_confidence=0.5
        ) as fd:
            results = fd.process(rgb)
    if not results.detections:
        return None
    best = max(results.detections, key=lambda d: d.score[0])
    bbox = best.location_data.relative_bounding_box
    x = int(bbox.xmin * w)
    y = int(bbox.ymin * h)
    width = int(bbox.width * w)
    height = int(bbox.height * h)
    x = max(0, min(x, w - 1))
    y = max(0, min(y, h - 1))
    width = max(1, min(width, w - x))
    height = max(1, min(height, h - y))
    return {
        "x": x,
        "y": y,
        "width": width,
        "height": height,
        "frame_width": w,
        "frame_height": h,
    }


def run_one_shot(path: Path | None, data: bytes | None) -> None:
    """One-shot mode: read image from path or data, print JSON and exit."""
    if path is not None:
        image = cv2.imread(str(path))
    else:
        if not data:
            print("{}", flush=True)
            sys.exit(1)
        nparr = np.frombuffer(data, np.uint8)
        image = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
    if image is None:
        print("{}", flush=True)
        sys.exit(1)
    out = detect_one(image)
    if out is None:
        print("{}", flush=True)
        sys.exit(1)
    print(json.dumps(out), flush=True)
    sys.exit(0)


# --- HTTP server (RPC) ---
_detector_lock = Lock()


def make_detect_handler(face_detection: "mp.solutions.face_detection.FaceDetection"):
    """Create a handler that reuses the given FaceDetection instance (model loaded once)."""

    class DetectHandler(BaseHTTPRequestHandler):
        def do_GET(self):
            if self.path == "/health":
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"ok":true}\n')
            else:
                self.send_response(404)
                self.end_headers()

        def do_POST(self):
            if self.path != "/detect":
                self.send_response(404)
                self.end_headers()
                return
            content_length = int(self.headers.get("Content-Length", 0))
            if content_length <= 0 or content_length > 10 * 1024 * 1024:  # 10MB
                self.send_response(400)
                self.end_headers()
                return
            body = self.rfile.read(content_length)
            nparr = np.frombuffer(body, np.uint8)
            image = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
            with _detector_lock:
                out = detect_one(image, face_detection)
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            if out is None:
                self.wfile.write(b"{}\n")
            else:
                self.wfile.write((json.dumps(out) + "\n").encode("utf-8"))

        def log_message(self, format, *args):
            pass  # quiet

    return DetectHandler


def run_server(port: int) -> None:
    with mp_face_detection.FaceDetection(
        model_selection=0, min_detection_confidence=0.5
    ) as face_detection:
        handler = make_detect_handler(face_detection)
        with HTTPServer(("127.0.0.1", port), handler) as httpd:
            print(f"face_detect_mediapipe: listening on 127.0.0.1:{port}", file=sys.stderr, flush=True)
            httpd.serve_forever()


def main() -> None:
    parser = argparse.ArgumentParser(description="MediaPipe face detection (one-shot or RPC server)")
    parser.add_argument("--serve", action="store_true", help="Run HTTP server for RPC")
    parser.add_argument("--port", type=int, default=8765, help="Server port (default 8765)")
    parser.add_argument("image_path", nargs="?", type=Path, help="Image file (one-shot mode)")
    args = parser.parse_args()

    if args.serve:
        if mp_face_detection is None:
            sys.exit(2)
        run_server(args.port)
        return

    # One-shot: image_path or stdin
    path = args.image_path if args.image_path and args.image_path.exists() else None
    data = None if path else sys.stdin.buffer.read()
    run_one_shot(path, data)


if __name__ == "__main__":
    main()

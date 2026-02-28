# desk-squat-tracker

スタンディングデスクでのスクワット検知アプリ。顔検出に MediaPipe（Python）を使用しています。

## 開発

### 前提条件

- **Go**: 1.22+
- **Python**: 3.12+（[uv](https://docs.astral.sh/uv/) 推奨）
- **asdf**（任意）: `.tool-versions` で uv / Python を管理

### Python ヘルパー（顔検出）

顔検出は `helper/` の Python スクリプトが担当し、Go アプリから HTTP で呼び出されます。

#### セットアップ

```bash
cd helper
uv sync
```

## Third-party licenses

This project uses the following third-party software.

### MediaPipe

Face detection library.  
https://github.com/google-ai-edge/mediapipe

```
Copyright 2020 The MediaPipe Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

### OpenCV

Image processing (opencv-python-headless).  
https://github.com/opencv/opencv-python

```
Copyright (C) 2000-2024, Intel Corporation, all rights reserved.
Copyright (C) 2009-2011, Willow Garage Inc., all rights reserved.
(and other contributors - see https://github.com/opencv/opencv)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

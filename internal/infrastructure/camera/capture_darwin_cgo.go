//go:build darwin

package camera

/*
#cgo darwin CFLAGS: -x objective-c -fobjc-arc -fmodules
#cgo darwin LDFLAGS: -framework AVFoundation -framework CoreMedia -framework CoreVideo -framework Foundation

#import <AVFoundation/AVFoundation.h>
#import <CoreMedia/CoreMedia.h>
#import <CoreVideo/CoreVideo.h>
#import <Foundation/Foundation.h>

#define CIF_WIDTH  352
#define CIF_HEIGHT 288

// キャプチャ用グローバル。デリゲートは gQueue（直列）で動作し、gLock で gFrameBuf/gFrameReady を保護。
// Go からは C を呼ぶのみ（C→Go コールバックなし）なので CGO のスレッド制約に注意不要。
static AVCaptureSession *gSession;
static dispatch_queue_t gQueue;
static uint8_t *gFrameBuf;
static int gFrameWidth;
static int gFrameHeight;
static int gFrameReady;
static NSLock *gLock;

static inline uint8_t clampByte(int v) {
	if (v < 0) return 0;
	if (v > 255) return 255;
	return (uint8_t)v;
}

// --- サンプルバッファ → YCbCr444 変換デリゲート ---
@interface DeskSquatFrameDelegate : NSObject <AVCaptureVideoDataOutputSampleBufferDelegate>
@end

@implementation DeskSquatFrameDelegate
- (void)captureOutput:(AVCaptureOutput *)output
didOutputSampleBuffer:(CMSampleBufferRef)sampleBuffer
       fromConnection:(AVCaptureConnection *)connection
{
	CVImageBufferRef img = CMSampleBufferGetImageBuffer(sampleBuffer);
	if (!img) return;
	CVPixelBufferLockBaseAddress(img, kCVPixelBufferLock_ReadOnly);
	size_t w = CVPixelBufferGetWidth(img);
	size_t h = CVPixelBufferGetHeight(img);
	OSType fmt = CVPixelBufferGetPixelFormatType(img);
	if (w == 0 || h == 0) {
		CVPixelBufferUnlockBaseAddress(img, kCVPixelBufferLock_ReadOnly);
		return;
	}
	size_t dstW = w, dstH = h, srcX0 = 0, srcY0 = 0, srcW = w, srcH = h;
	if (w > CIF_WIDTH || h > CIF_HEIGHT) {
		double dstAspect = (double)CIF_WIDTH / (double)CIF_HEIGHT;
		double srcAspect = (double)w / (double)h;
		dstW = CIF_WIDTH;
		dstH = CIF_HEIGHT;
		if (srcAspect > dstAspect) {
			srcH = h;
			srcW = (size_t)((double)h * dstAspect);
			if (srcW > w) srcW = w;
			srcX0 = (w > srcW) ? (w - srcW) / 2 : 0;
			srcY0 = 0;
		} else {
			srcW = w;
			srcH = (size_t)((double)w / dstAspect);
			if (srcH > h) srcH = h;
			srcX0 = 0;
			srcY0 = (h > srcH) ? (h - srcH) / 2 : 0;
		}
	}
	size_t bufSize444 = dstW * dstH * 3;

	if (fmt == kCVPixelFormatType_420YpCbCr8BiPlanarFullRange || fmt == kCVPixelFormatType_420YpCbCr8BiPlanarVideoRange) {
		if (!CVPixelBufferIsPlanar(img) || CVPixelBufferGetPlaneCount(img) < 2) {
			CVPixelBufferUnlockBaseAddress(img, kCVPixelBufferLock_ReadOnly);
			return;
		}
		uint8_t *srcY = (uint8_t *)CVPixelBufferGetBaseAddressOfPlane(img, 0);
		size_t strideY = CVPixelBufferGetBytesPerRowOfPlane(img, 0);
		uint8_t *srcUV = (uint8_t *)CVPixelBufferGetBaseAddressOfPlane(img, 1);
		size_t strideUV = CVPixelBufferGetBytesPerRowOfPlane(img, 1);
		if (!srcY || !srcUV || strideY == 0 || strideUV == 0) {
			CVPixelBufferUnlockBaseAddress(img, kCVPixelBufferLock_ReadOnly);
			return;
		}
		[gLock lock];
		if (!gFrameBuf || gFrameWidth != (int)dstW || gFrameHeight != (int)dstH) {
			if (gFrameBuf) free(gFrameBuf);
			gFrameBuf = (uint8_t *)malloc(bufSize444);
			gFrameWidth = (int)dstW;
			gFrameHeight = (int)dstH;
		}
		if (gFrameBuf) {
			uint8_t *dst = gFrameBuf;
			for (size_t dy = 0; dy < dstH; dy++) {
				size_t syRel = (dstH > 0) ? (dy * srcH) / dstH : 0;
				size_t sy = srcY0 + syRel;
				if (sy >= h) sy = h - 1;
				uint8_t *rowY = srcY + sy * strideY;
				uint8_t *rowUV = srcUV + (sy / 2) * strideUV;
				for (size_t dx = 0; dx < dstW; dx++) {
					size_t sxRel = (dstW > 0) ? (dx * srcW) / dstW : 0;
					size_t sx = srcX0 + sxRel;
					if (sx >= w) sx = w - 1;
					uint8_t Y = rowY[sx];
					size_t uvx = (sx / 2) * 2;
					uint8_t Cb = rowUV[uvx];
					uint8_t Cr = rowUV[uvx + 1];
					size_t di = (dy * dstW + dx) * 3;
					dst[di] = Y;
					dst[di+1] = Cb;
					dst[di+2] = Cr;
				}
			}
			gFrameReady = 1;
		}
		[gLock unlock];
		CVPixelBufferUnlockBaseAddress(img, kCVPixelBufferLock_ReadOnly);
		return;
	}

	if (fmt == kCVPixelFormatType_32BGRA) {
		uint8_t *src = (uint8_t *)CVPixelBufferGetBaseAddress(img);
		size_t stride = CVPixelBufferGetBytesPerRow(img);
		if (!src || stride == 0) {
			CVPixelBufferUnlockBaseAddress(img, kCVPixelBufferLock_ReadOnly);
			return;
		}
		[gLock lock];
		if (!gFrameBuf || gFrameWidth != (int)dstW || gFrameHeight != (int)dstH) {
			if (gFrameBuf) free(gFrameBuf);
			gFrameBuf = (uint8_t *)malloc(bufSize444);
			gFrameWidth = (int)dstW;
			gFrameHeight = (int)dstH;
		}
		if (gFrameBuf) {
			uint8_t *dst = gFrameBuf;
			for (size_t dy = 0; dy < dstH; dy++) {
				size_t syRel = (dstH > 0) ? (dy * srcH) / dstH : 0;
				size_t sy = srcY0 + syRel;
				if (sy >= h) sy = h - 1;
				uint8_t *row = src + sy * stride;
				for (size_t dx = 0; dx < dstW; dx++) {
					size_t sxRel = (dstW > 0) ? (dx * srcW) / dstW : 0;
					size_t sx = srcX0 + sxRel;
					if (sx >= w) sx = w - 1;
					size_t si = sx * 4;
					int iR = (int)row[si+2], iG = (int)row[si+1], iB = (int)row[si];
					int Y = (66*iR + 129*iG + 25*iB + 128) >> 8; Y += 16;
					int Cb = (-38*iR - 74*iG + 112*iB + 128) >> 8; Cb += 128;
					int Cr = (112*iR - 94*iG - 18*iB + 128) >> 8; Cr += 128;
					size_t di = (dy * dstW + dx) * 3;
					dst[di] = clampByte(Y);
					dst[di+1] = clampByte(Cb);
					dst[di+2] = clampByte(Cr);
				}
			}
			gFrameReady = 1;
		}
		[gLock unlock];
	}
	CVPixelBufferUnlockBaseAddress(img, kCVPixelBufferLock_ReadOnly);
}
@end

static DeskSquatFrameDelegate *gDelegate;

// --- デバイス一覧（List / StartCapture で同じ順序を保つため共通）---
static NSArray *getVideoDevices(void) {
	AVCaptureDeviceDiscoverySession *d = [AVCaptureDeviceDiscoverySession discoverySessionWithDeviceTypes:@[AVCaptureDeviceTypeBuiltInWideAngleCamera, AVCaptureDeviceTypeExternalUnknown] mediaType:AVMediaTypeVideo position:AVCaptureDevicePositionUnspecified];
	return d.devices;
}

int ListCameraCount(void) {
	@autoreleasepool {
		return (int)[getVideoDevices() count];
	}
}
int ListCameraNameAtIndex(int index, char *buf, int bufLen) {
	if (!buf || bufLen <= 0) return -1;
	@autoreleasepool {
		NSArray *devices = getVideoDevices();
		if (index < 0 || index >= (int)[devices count]) return -1;
		AVCaptureDevice *dev = [devices objectAtIndex:(NSUInteger)index];
		NSString *name = dev.localizedName;
		if (!name) return -1;
		NSData *data = [name dataUsingEncoding:NSUTF8StringEncoding];
		if (!data || [data length] >= (NSUInteger)bufLen) return -1;
		memcpy(buf, [data bytes], [data length]);
		buf[[data length]] = '\0';
		return (int)[data length];
	}
}

// StartCaptureWithDeviceIndex: 0 ok, <0 error. device_index でカメラを指定。
int StartCaptureWithDeviceIndex(int device_index) {
	@autoreleasepool {
		gLock = [NSLock new];
		NSArray *devices = getVideoDevices();
		if (!devices || device_index < 0 || device_index >= (int)[devices count]) return -1;
		AVCaptureDevice *dev = [devices objectAtIndex:(NSUInteger)device_index];
		if (!dev) return -1;

		NSError *err = nil;
		AVCaptureDeviceInput *input = [AVCaptureDeviceInput deviceInputWithDevice:dev error:&err];
		if (err || !input) return -2;

		AVCaptureSession *session = [[AVCaptureSession alloc] init];
		if (!session) return -3;

		[session beginConfiguration];
		if ([session canSetSessionPreset:AVCaptureSessionPreset352x288]) {
			session.sessionPreset = AVCaptureSessionPreset352x288;
		}
		gFrameWidth = 352;
		gFrameHeight = 288;

		if (![session canAddInput:input]) return -4;
		[session addInput:input];

		// デリゲート側で変換対応している 420BiPlanar のみを明示指定（444 は未対応のため選ばない）
		AVCaptureVideoDataOutput *out = [[AVCaptureVideoDataOutput alloc] init];
		NSDictionary *settings = @{ (id)kCVPixelBufferPixelFormatTypeKey : @(kCVPixelFormatType_420YpCbCr8BiPlanarFullRange) };
		out.videoSettings = settings;
		out.alwaysDiscardsLateVideoFrames = YES;

		gDelegate = [DeskSquatFrameDelegate new];
		gQueue = dispatch_queue_create("go.av.capture", DISPATCH_QUEUE_SERIAL);
		[out setSampleBufferDelegate:gDelegate queue:gQueue];

		if (![session canAddOutput:out]) return -5;
		[session addOutput:out];
		[session commitConfiguration];
		[session startRunning];
		gSession = session;
	}
	return 0;
}

void StopCaptureDarwin(void) {
	@autoreleasepool {
		if (gSession) {
			[gSession stopRunning];
			gSession = nil;
		}
		// デリゲートが gLock を保持したまま gFrameBuf に書き込む可能性があるため、
		// 解放前に gLock を取得してコールバック完了を待つ。
		NSLock *lock = gLock;
		if (lock) {
			[lock lock];
		}
		if (gFrameBuf) {
			free(gFrameBuf);
			gFrameBuf = NULL;
		}
		gFrameWidth = 0;
		gFrameHeight = 0;
		gFrameReady = 0;
		gDelegate = nil;
		gQueue = nil;
		if (lock) {
			[lock unlock];
			gLock = nil;
		}
	}
}

// GetFrameDarwin は最新フレームを取得。*buf は C 側の gFrameBuf を指すので、Go 側で GoBytes でコピーすること。
int GetFrameDarwin(uint8_t **buf, int *w, int *h, int *frameSizeOut) {
	if (!gFrameBuf || !gLock) return -1;
	[gLock lock];
	if (!gFrameReady) {
		[gLock unlock];
		return -1;
	}
	*buf = gFrameBuf;
	*w = gFrameWidth;
	*h = gFrameHeight;
	if (frameSizeOut) *frameSizeOut = gFrameWidth * gFrameHeight * 3;
	gFrameReady = 0;
	[gLock unlock];
	return 0;
}
*/
import "C"

import (
	"context"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/xerrors"
)

func startStreamDarwinCGO(ctx context.Context, deviceIndex int) (<-chan Frame, error) {
	rc := C.StartCaptureWithDeviceIndex(C.int(deviceIndex))
	if rc != 0 {
		return nil, xerrors.Errorf("cannot start capture, rc=%d", int(rc))
	}
	out := make(chan Frame, 1)
	go runDarwinCaptureLoop(ctx, out)
	return out, nil
}

// runDarwinCaptureLoop は C のキャプチャを止めずにフレームを out へ送る。ctx で終了。
func runDarwinCaptureLoop(ctx context.Context, out chan<- Frame) {
	defer close(out)
	defer C.StopCaptureDarwin()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		var cbuf *C.uchar
		var cw, ch, csize C.int
		if C.GetFrameDarwin(&cbuf, &cw, &ch, &csize) != 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		w, h, size := int(cw), int(ch), int(csize)
		if w <= 0 || h <= 0 || size <= 0 || cbuf == nil || size != w*h*3 {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		data := C.GoBytes(unsafe.Pointer(cbuf), C.int(size))
		select {
		case out <- Frame{Data: data, Width: w, Height: h}:
		case <-ctx.Done():
			return
		default:
			// ドロップ
		}
	}
}

func listDevicesDarwin() ([]Device, error) {
	n := int(C.ListCameraCount())
	if n <= 0 {
		return listDevicesDefault()
	}
	devs := make([]Device, 0, n)
	for i := 0; i < n; i++ {
		name := darwinDeviceNameAt(i)
		if name == "" {
			name = "Camera " + strconv.Itoa(i)
		}
		devs = append(devs, Device{Index: i, Name: name, ID: ""})
	}
	return devs, nil
}

func darwinDeviceNameAt(index int) string {
	buf := make([]byte, 256)
	written := int(C.ListCameraNameAtIndex(C.int(index), (*C.char)(unsafe.Pointer(&buf[0])), C.int(len(buf))))
	if written <= 0 || written >= len(buf) {
		return ""
	}
	return strings.TrimSpace(string(buf[:written]))
}

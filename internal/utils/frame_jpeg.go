package utils

import (
	"bytes"
	"image"
	"image/jpeg"
)

// Frame is a single frame from the camera (packed YCbCr444).
// Data length = Width * Height * 3.
type Frame struct {
	Data   []byte
	Width  int
	Height int
}

// EncodeJPEG converts a YCbCr444-packed frame to JPEG bytes.
// Quality is 1-100; typical value 70-85.
func EncodeJPEG(f Frame, quality int) ([]byte, error) {
	if len(f.Data) != f.Width*f.Height*3 || f.Width <= 0 || f.Height <= 0 {
		return nil, nil
	}
	img := packedYCbCr444ToImage(f)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// packedYCbCr444ToImage converts packed YCbCr444 (Y,Cb,Cr per pixel) to image.YCbCr.
func packedYCbCr444ToImage(f Frame) *image.YCbCr {
	n := f.Width * f.Height
	y := make([]byte, n)
	cb := make([]byte, n)
	cr := make([]byte, n)
	for i := 0; i < n; i++ {
		y[i] = f.Data[i*3]
		cb[i] = f.Data[i*3+1]
		cr[i] = f.Data[i*3+2]
	}
	return &image.YCbCr{
		Y:              y,
		Cb:             cb,
		Cr:             cr,
		YStride:        f.Width,
		CStride:        f.Width,
		SubsampleRatio: image.YCbCrSubsampleRatio444,
		Rect:           image.Rect(0, 0, f.Width, f.Height),
	}
}

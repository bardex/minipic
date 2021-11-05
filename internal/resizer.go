package internal

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"math"

	"github.com/disintegration/imaging"
)

type ResizerByImaging struct{}

func (r ResizerByImaging) Resize(src io.Reader, dst io.Writer, opts ResizeOptions) error {
	img, imtype, err := image.Decode(src)
	if err != nil {
		return fmt.Errorf("image decode error:%w", err)
	}

	srcWidth := float64(img.Bounds().Dx())
	srcHeight := float64(img.Bounds().Dy())

	switch opts.Mode {
	case ResizeModeFit:
		k := math.Max(srcWidth/float64(opts.Width), srcHeight/float64(opts.Height))
		width := int(math.Round(srcWidth / k))
		height := int(math.Round(srcHeight / k))
		img = imaging.Resize(img, width, height, imaging.Lanczos)
	case ResizeModeFill:
		k := math.Min(srcWidth/float64(opts.Width), srcHeight/float64(opts.Height))
		width := int(math.Round(srcWidth / k))
		height := int(math.Round(srcHeight / k))
		img = imaging.Resize(img, width, height, imaging.Lanczos)
		if width > opts.Width {
			padX := (width - opts.Width) / 2
			img = imaging.Crop(img, image.Rect(padX, 0, padX+opts.Width, height))
		}
		if height > opts.Height {
			padY := (height - opts.Height) / 2
			img = imaging.Crop(img, image.Rect(0, padY, width, padY+opts.Height))
		}
	default:
		return fmt.Errorf("unknown resize mode: %s", opts.Mode)
	}

	switch imtype {
	case "jpeg":
		err = imaging.Encode(dst, img, imaging.JPEG, imaging.JPEGQuality(85))
		if err != nil {
			return fmt.Errorf("image encode error:%w", err)
		}
	case "png":
		err = imaging.Encode(dst, img, imaging.PNG, imaging.PNGCompressionLevel(png.BestCompression))
		if err != nil {
			return fmt.Errorf("image encode error:%w", err)
		}
	default:
		return fmt.Errorf("image encode error: unknown type:%s", imtype)
	}

	return nil
}

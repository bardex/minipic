package app

import (
	"errors"
	"fmt"
	"image"

	// init jpeg decoder.
	_ "image/jpeg"
	"image/png"
	"io"
	"math"

	"github.com/bardex/minipic/internal/httpserver"
	"github.com/disintegration/imaging"
)

var (
	// ErrUnsupportedFormat unsupported image format.
	ErrUnsupportedFormat = errors.New("unsupported image format")
	// ErrUnsupportedMode unsupported resize mode.
	ErrUnsupportedMode = errors.New("unsupported resize mode")
)

type Resizer struct{}

func (r Resizer) Resize(src io.Reader, dst io.Writer, opts httpserver.ResizeOptions) error {
	img, imtype, err := image.Decode(src)
	if err != nil {
		if errors.Is(err, image.ErrFormat) {
			return ErrUnsupportedFormat
		}
		return err
	}

	srcWidth := float64(img.Bounds().Dx())
	srcHeight := float64(img.Bounds().Dy())

	switch opts.Mode {
	case httpserver.ResizeModeFit:
		k := math.Max(srcWidth/float64(opts.Width), srcHeight/float64(opts.Height))
		width := int(math.Round(srcWidth / k))
		height := int(math.Round(srcHeight / k))
		img = imaging.Resize(img, width, height, imaging.Lanczos)
	case httpserver.ResizeModeFill:
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
		return fmt.Errorf("%w: %s", ErrUnsupportedMode, opts.Mode)
	}

	switch imtype {
	case "jpeg":
		err = imaging.Encode(dst, img, imaging.JPEG, imaging.JPEGQuality(85))
		if err != nil {
			if errors.Is(err, image.ErrFormat) {
				return ErrUnsupportedFormat
			}
			return err
		}
	case "png":
		err = imaging.Encode(dst, img, imaging.PNG, imaging.PNGCompressionLevel(png.BestCompression))
		if err != nil {
			if errors.Is(err, image.ErrFormat) {
				return ErrUnsupportedFormat
			}
			return err
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedFormat, imtype)
	}

	return nil
}

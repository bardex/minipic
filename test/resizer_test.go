package test

import (
	"fmt"
	"image"
	"os"
	"testing"

	"github.com/bardex/minipic/internal/app"
	"github.com/bardex/minipic/internal/httpserver"
	"github.com/stretchr/testify/require"
)

func TestResizer(t *testing.T) {
	tests := []struct {
		file   string
		mode   string
		width  int
		height int
		format string
		err    error
	}{
		{file: "sample.jpeg", mode: "fill", width: 800, height: 600, format: "jpeg", err: nil},
		{file: "sample.jpeg", mode: "fit", width: 800, height: 800, format: "jpeg", err: nil},
		{file: "sample.png", mode: "fill", width: 800, height: 600, format: "png", err: nil},
		{file: "sample.png", mode: "fit", width: 800, height: 800, format: "png", err: nil},
		{file: "sample_v.png", mode: "fill", width: 800, height: 600, format: "png", err: nil},
		{file: "sample_v.png", mode: "fit", width: 800, height: 800, format: "png", err: nil},
		{file: "sample_v.png", mode: "crop", width: 800, height: 800, format: "png", err: app.ErrUnsupportedMode},
		{file: "sample.webp", mode: "fill", width: 600, height: 800, format: "", err: app.ErrUnsupportedFormat},
		{file: "sample.webp", mode: "fit", width: 800, height: 800, format: "", err: app.ErrUnsupportedFormat},
	}
	resizer := app.Resizer{}
	for _, tt := range tests {
		tt := tt
		name := fmt.Sprintf("%s_%s_%d_%d", tt.file, tt.mode, tt.width, tt.height)
		t.Run(name, func(t *testing.T) {
			src, err := os.Open(tt.file)
			require.NoError(t, err)
			defer src.Close()

			dst, err := os.CreateTemp("../tmp", "minipictest")
			require.NoError(t, err)
			defer dst.Close()
			defer os.Remove(dst.Name())

			err = resizer.Resize(src, dst, httpserver.ResizeOptions{Mode: tt.mode, Width: tt.width, Height: tt.height})
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			_, err = dst.Seek(0, 0)
			require.NoError(t, err)

			img, format, err := image.Decode(dst)
			require.NoError(t, err)

			w := img.Bounds().Max.X
			h := img.Bounds().Max.Y

			switch tt.mode {
			case "fill":
				require.Equal(t, tt.width, w)
				require.Equal(t, tt.height, h)
			case "fit":
				require.LessOrEqual(t, w, tt.width)
				require.LessOrEqual(t, h, tt.height)
			default:
				t.Fatalf("unknown resize mode: %s", tt.mode)
			}
			require.Equal(t, tt.format, format)
		})
	}
}

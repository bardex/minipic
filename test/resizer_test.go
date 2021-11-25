package test

import (
	"os"
	"testing"

	"github.com/bardex/minipic/internal/app"
	"github.com/bardex/minipic/internal/httpserver"
	"github.com/stretchr/testify/require"
)

func TestResizerByImaging_Resize(t *testing.T) {
	var resizer httpserver.ImageResizer = app.Resizer{}

	src, err := os.Open("sample.jpg")
	require.NoError(t, err)
	defer src.Close()

	dst, err := os.CreateTemp("/tmp", "mp_resize_")
	require.NoError(t, err)
	defer dst.Close()
	defer os.Remove(dst.Name())

	err = resizer.Resize(src, dst, httpserver.ResizeOptions{Mode: "fill", Width: 600, Height: 600})
	require.NoError(t, err)
}

package lisgo

import (
	"image"
	"image/color"
)

type ImageBmpGray struct {
	data     []byte
	header   *BmpHeader
	scanLine int
}

func NewBmpGrayImage(data []byte, header *BmpHeader) *ImageBmpGray{
	//it seems that we may ignore palette
	img := ImageBmpGray{
		data:     data,
		header:   header,
		scanLine: int(header.Width + pad4(header.Width)),
	}
	return &img
}

// ColorModel returns the Image's color model.
func (i *ImageBmpGray) ColorModel() color.Model {
	return color.GrayModel
}

// Bounds returns the domain for which At can return non-zero color.
// The bounds do not necessarily contain the point (0, 0).
func (i *ImageBmpGray) Bounds() image.Rectangle {
	return image.Rect(0, 0, int(i.header.Width), int(abs(i.header.Height)))
}

// At returns the color of the pixel at (x, y).
func (i *ImageBmpGray) At(x, y int) color.Color {

	//x = int(i.header.Width) - x - 1
	if i.header.Height > 0 {
		y = int(abs(i.header.Height)) - y - 1
	}

	offset := i.scanLine* y + x
	return color.Gray{ Y: i.data[offset] }
}



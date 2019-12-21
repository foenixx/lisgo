package lisgo

import (
	"github.com/apex/log"
	"image"
	"image/color"
)

//ImageBmpBw represents a black-N-white image with 1 bit per pixel
type ImageBmpBw struct {
	data     []byte
	header   *BmpHeader
	scanLine int
	palette  color.Palette
}

//NewBmpBwImage constructs new object
func NewBmpBwImage(data []byte, header *BmpHeader) *ImageBmpBw {
	img := ImageBmpBw{
		data:     data,
		header:   header,
		scanLine: int(pad4(header.Width/8)),
		palette: color.Palette{
			color.RGBA{R: 0, G: 0, B: 0, A: 255},        //black
			color.RGBA{R: 255, G: 255, B: 255, A: 255}}, //white
	}
	log.WithField("scanLine", img.scanLine).Debug("creating ImgBmpBw")
	return &img
}

// ColorModel returns the Image's color model.
func (i *ImageBmpBw) ColorModel() color.Model {
	return &i.palette
}

// Bounds returns the domain for which At can return non-zero color.
// The bounds do not necessarily contain the point (0, 0).
func (i *ImageBmpBw) Bounds() image.Rectangle {
	return image.Rect(0, 0, int(i.header.Width), int(abs(i.header.Height)))
}

//ColorIndexAt returns the color index of the pixel at (x, y).
func (i *ImageBmpBw) ColorIndexAt(x, y int) uint8 {
	if i.header.Height > 0 {
		//bottom-up image
		y = int(i.header.Height) - y - 1
	}

	byteIndex := x / 8
	bitIndex := 7 - x%8

	offset := y*i.scanLine + byteIndex
	var mask uint8 = 1 << bitIndex
	if (i.data[offset] & mask) == mask {
		return 1
	} else {
		return 0
	}
}

// At returns the color of the pixel at (x, y).
func (i *ImageBmpBw) At(x, y int) color.Color {
	return i.palette[i.ColorIndexAt(x, y)]
}

package lisgo

import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/apex/log"
	"io"
)

// https://www.fileformat.info/format/bmp/egff.htm

//BmpHeader represents a BMP version 3 header (54 bytes total)
type BmpHeader struct {
	Magic        uint16
	FileSize     uint32
	Unused       uint32
	OffsetToData uint32
	HeaderSize   uint32
	//Width and Height are the width and height of the image in pixels, respectively. If Height is a positive number,
	//then the image is a "bottom-up" bitmap with the origin in the lower-left corner. If Height is a negative number,
	//then the image is a "top-down" bitmap with the origin in the upper-left corner. Width does not include any scan-line boundary padding.
	Width                uint32
	Height               int32
	NbColorPlanes        uint16
	NbBitsPerPixel       uint16
	Compression          uint32
	PixelDataSize        uint32
	HorizontalResolution uint32
	VerticalResolution   uint32
	NbColorsInPalette    uint32
	ImportantColors      uint32
}

const (
	bmp2HeaderSize = 14
	bmp3HeaderSize = 40
	bmpHeaderSize  = bmp2HeaderSize + bmp3HeaderSize
)

func ReadBMPHeader(r io.Reader) (*BmpHeader, []byte, error) {

	buf := make([]byte, bmpHeaderSize)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, nil, err
	}
	var header BmpHeader
	err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, &header)
	if err != nil {
		return nil, nil, err
	}
	log.WithField("header", fmt.Sprintf("%+v\n", header)).Debug("header is read")
	if header.HeaderSize == bmp3HeaderSize {
		paletteSize := header.OffsetToData - (bmp2HeaderSize + bmp3HeaderSize)
		if paletteSize > 0 {
			log.WithField("palette size", paletteSize).Debug("palette detected")
			palBuf := make([]byte, paletteSize)
			_, err := io.ReadFull(r, palBuf)
			if err != nil {
				return nil, nil, err
			}
			buf = append(buf, palBuf...)
		}
	}
	return &header, buf, nil
}

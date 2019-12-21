package lisgo

import "C"
import (
	"bytes"
	"errors"
	"github.com/apex/log"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	"golang.org/x/image/bmp"
)

//PageReader represents a single page received from scanner
type PageReader struct {
	Width          int
	Height         int
	Format         uint32
	Session        *ScanSession
	internalBuffer []byte //a byte array from C-code, read-only
	readBytes      int    //count of bytes read from internalBuffer, if equal to len(internalbuffer) then the buffer is completely read
}

//Read portion of data from scanner into a buffer until the page is over. Optimal size of the buffer is ScanSessionCBufferSize.
func (sb *PageReader) Read(p []byte) (int, error) {
	var err error
	var bts int
	bufcap := len(p)

	//fmt.Printf("read: bufcap: %d, readBytes: %d, len(internalBuffer): %d\n", bufcap, sb.readBytes, len(sb.internalBuffer))
	if bufcap == 0 {
		return 0, nil //nothing happened
	}

	//there are unread data in the internal buffer
	if sb.readBytes > 0 && sb.readBytes < len(sb.internalBuffer) {
		rest := len(sb.internalBuffer) - sb.readBytes

		if rest > bufcap {
			bts = bufcap
		} else {
			bts = rest
		}
		copy(p, sb.internalBuffer[sb.readBytes:sb.readBytes+bts])
		sb.readBytes += bts
		return bts, nil
	}

	// readBytes is 0 or len(internalBuffer) - first call or no unread data in the buffer left
	sb.readBytes = 0

	if sb.Session.EndOfPage() {
		return 0, io.EOF
	}

	var got uint64
	sb.internalBuffer, got, err = sb.Session.ScanRead()
	if err != nil {
		return int(got), err
	}

	//fmt.Print("#")

	//and again we have some unread data in the internal buffer with length=got

	if int(got) > bufcap {
		bts = bufcap
	} else {
		bts = int(got)
	}

	copy(p, sb.internalBuffer[:bts])
	sb.readBytes += bts
	return bts, nil
}

func (sb *PageReader) GetImage() (image.Image, error) {

	log.Debug("reading header")
	header, buf, err := ReadBMPHeader(sb)
	if err != nil {
		return nil, err
	}

	if (header.NbBitsPerPixel == 1 || header.NbBitsPerPixel == 8) && header.Compression == 0 {
		// White and black or Grayscale image without compression
		data := bytes.Buffer{}
		// height could be negative, that means image starts from top-left corner instead of bottom-right
		data.Grow(int(header.PixelDataSize))
		log.Debug("reading image data")
		n, err := data.ReadFrom(sb)
		if err != nil {
			return nil, err
		}
		log.WithField("bytes", n).Debug("image data is read")
		if header.NbBitsPerPixel == 1 {
			img := NewBmpBwImage(data.Bytes(), header)
			return img, nil
		}
		img := NewBmpGrayImage(data.Bytes(), header)
		return img, nil
	}

	log.Debug("reading image data")
	// pass it to the standard bmp decoder
	img, err := bmp.Decode(io.MultiReader(bytes.NewReader(buf), sb))
	log.Debug("image data is read")
	return img, err
}

func (sb *PageReader) WriteToFile(name string, format string) error {
	outputFile, err := os.Create(name)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	img, err := sb.GetImage()
	if err != nil {
		return err
	}
	switch format {
	case "png":
		return png.Encode(outputFile, img)
	case "jpg":
		return jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 50})
	}
	return errors.New("unknown file format")

}

//WriteToPng writes image to file
func (sb *PageReader) WriteToPng(name string) error {
	return sb.WriteToFile(name, "png")
}

//WriteToJpeg writes image to file
func (sb *PageReader) WriteToJpeg(name string) error {
	return sb.WriteToFile(name, "jpg")
}

//NewPageReader converts data buffer to image object
func NewPageReader(session *ScanSession, param *ScanParameters) *PageReader {

	b := PageReader{
		Width:   param.Width(),
		Height:  param.Height(),
		Format:  param.ImageFormat(),
		Session: session,
	}

	return &b

}

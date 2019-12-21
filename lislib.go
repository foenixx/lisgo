package lisgo

// LDFLAGS: --static ${SRCDIR}/../libinsane.a -lregex -lole32  -loleaut32 -luuid -lsystre -ltre  -lpthread -lintl -liconv

/*
#include <libinsane/capi.h>
#include <lislib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/mattn/go-pointer"
)

type (
	//lisgo is a holder for *lis_api
	lisgo struct {
		lisgo *C.struct_lis_api
	}

	//Scanner is a descriptor of scanner
	Scanner struct {
		lisDevice *C.struct_lis_item
		lis       *lisgo
		DeviceID  string
		Vendor    string
		Model     string
		Type      string
	}

	//ScanParameters holds the current scan session's parameters
	ScanParameters struct {
		lisScanParameters C.struct_lis_scan_parameters
	}
)

//mostly buffer sizes here
const (
	ScanSessionCBufferSize = 1024 * 1024 //1MB
)

//New creates new instances of lis_api using lis_safe_bet
func New() (*lisgo, error) {
	C.set_log_callbacks()
	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)

	var lib lisgo
	lib.lisgo = C.lis_api_get_api(errProxy.GetProxy())
	if errProxy.ErrNum() != LisOk {
		err := errors.New(errProxy.Error())
		return nil, err
	}
	return &lib, nil
}

//Close releases lis_api and and all connected objects
func (o *lisgo) Close() {
	C.lis_api_cleanup_proxy(o.lisgo)
}

//ListDevices returns available scanners (online for WIA and all for Twain)
func (o *lisgo) ListDevices() ([]*Scanner, error) {
	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)

	cdevs := C.lis_api_list_devices_proxy(o.lisgo, errProxy.GetProxy())
	if errProxy.ErrNum() != LisOk {
		err := errors.New(errProxy.Error())
		return nil, err
	}

	arrLen := C.lis_array_length(unsafe.Pointer(cdevs))

	//Apply strong magic to get GO slice backed by C null-terminated array
	devs := (*[maxSliceLen]*C.struct_lis_device_descriptor)(unsafe.Pointer(cdevs))[:arrLen:arrLen]

	devices := make([]*Scanner, 0)
	for _, d := range devs {
		devices = append(devices, d.conv2Go(o))
	}

	return devices, nil

}

//GetDevice searches thru the list of scanners for a device with specified id
func (o *lisgo) GetDevice(deviceID string) (*Scanner, error) {

	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)
	cdevs := C.lis_api_list_devices_proxy(o.lisgo, errProxy.GetProxy())
	if errProxy.ErrNum() != LisOk {
		err := errors.New(errProxy.Error())
		return nil, err
	}

	arrlen := C.lis_array_length(unsafe.Pointer(cdevs))

	//Apply strong magic to get GO slice backed by C null-terminated array
	devs := (*[maxSliceLen]*C.struct_lis_device_descriptor)(unsafe.Pointer(cdevs))[:arrlen:arrlen]

	for _, d := range devs {
		dev := d.conv2Go(o)
		if dev.DeviceID == deviceID {
			return dev, nil
		}
	}

	return nil, nil

}

func (d *C.struct_lis_device_descriptor) conv2Go(o *lisgo) *Scanner {
	return &Scanner{
		nil,
		o,
		C.GoString(d.dev_id),
		C.GoString(d.vendor),
		C.GoString(d.model),
		C.GoString(d._type),
	}
}

//PaperSource represents a source of paper for scan, i.e flatbed or automatic feeder
type PaperSource struct {
	Name   string
	Kind   C.enum_lis_item_type
	source *C.struct_lis_item
}

type iterSourcesCallback struct {
	callback func(*PaperSource) bool
}

type iterOptionsCallback struct {
	callback func(*OptionDescriptor) bool
}

//export iterSourcesProxy
func iterSourcesProxy(cb unsafe.Pointer, sourcePtr *C.struct_lis_item, sourceName *C.char, kind C.enum_lis_item_type) C.int {
	res := &PaperSource{
		Name:   C.GoString(sourceName),
		Kind:   kind,
		source: sourcePtr,
	}
	if pointer.Restore(cb).(*iterSourcesCallback).callback(res) {
		return 1
	}
	return 0
}

//export iterOptionsProxy
func iterOptionsProxy(cb unsafe.Pointer, opt *C.struct_lis_option_descriptor, valType C.enum_lis_value_type, conType C.int, conPossible unsafe.Pointer) C.int {


	res := &OptionDescriptor{
		optStruct:    opt,
		Name:         C.GoString(opt.name),
		Title:        C.GoString(opt.title),
		Desc:         C.GoString(opt.desc),
		Capabilities: int(opt.capabilities),
		ValueType:    valType,
		ValueUnit:    opt.value.unit,
		Constraint:   NewConstraint(valType, conType, conPossible),
	}

	if pointer.Restore(cb).(*iterOptionsCallback).callback(res) {
		return 1
	}
	return 0
}

//Open calls lis->get_device. Should be called before any of GetSourceByName, IterateOptions are called.
func (d *Scanner) Open() error {
	var dev *C.char = C.CString(d.DeviceID)
	defer C.free(unsafe.Pointer(dev))

	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)

	d.lisDevice = C.lis_api_get_device_proxy(d.lis.lisgo, dev, errProxy.GetProxy())
	if errProxy.ErrNum() != LisOk {
		return errors.New(errProxy.Error())
	}
	return nil
}

//Close should be called after Open to release associated resources
func (d *Scanner) Close() {
	C.lis_item_close_proxy(d.lisDevice)
	d.lisDevice = nil
}

//GetPaperSource returns paper source with specified name if any.
// Otherwise it returns nil.
func (d *Scanner) GetPaperSource(name string) (*PaperSource, error) {
	var source *PaperSource
	err := d.iterateSourcesInternal(nil, d.lisDevice, func(s *PaperSource) bool {
		if s.Name == name {
			source = s
			return false
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return source, nil
}

//IterateSources iterates thru paper sources
func (d *Scanner) IterateSources(f func(*PaperSource) bool) error {
	if d.lisDevice == nil {
		var dev *C.char = C.CString(d.DeviceID)
		defer C.free(unsafe.Pointer(dev))
		//device's gonna be open and then closed automatically
		return d.iterateSourcesInternal(dev, nil, f)
	} else {
		//device is already open
		//fmt.Printf("Device [%v] is already open\n", d.DeviceID)
		return d.iterateSourcesInternal(nil, d.lisDevice, f)
	}
}

func (d *Scanner) iterateSourcesInternal(devID *C.char, devRef *C.struct_lis_item, f func(*PaperSource) bool) error {
	proxy := iterSourcesCallback{f}
	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)

	C.lis_item_iterate_sources(d.lis.lisgo, devID, devRef, pointer.Save(&proxy), errProxy.GetProxy())
	if errProxy.ErrNum() != LisOk {
		return errors.New(errProxy.Error())
	}
	return nil
}

//IterateOptions iterates thru options of the paper source
func (s *PaperSource) IterateOptions(f func(*OptionDescriptor) bool) error {
	proxy := iterOptionsCallback{f}

	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)

	C.lis_item_iterate_options(s.source, pointer.Save(&proxy), errProxy.GetProxy())
	if errProxy.ErrNum() != LisOk {
		return errors.New(errProxy.Error())
	}
	return nil
}

//SetOption accepts string representation of value, converts it to the actual type of the option and sets it.
func (s *PaperSource) SetOption(name string, val string) error {
	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cVal := C.CString(val)
	defer C.free(unsafe.Pointer(cVal))
	C.lis_set_option_proxy(s.source, cName, cVal, errProxy.GetProxy())
	if errProxy.ErrNum() != LisOk {
		return errors.New(errProxy.Error())
	}
	return nil
}

//ScanStart creates scanning session
func (s *PaperSource) ScanStart() (*ScanSession, error) {
	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)
	var err error
	lisSession := C.lis_item_scan_start_proxy(s.source, errProxy.GetProxy())
	if errProxy.ErrNum() != LisOk {
		err = errors.New(errProxy.Error())
		return nil, err
	}

	session := ScanSession{
		lisScanSession: lisSession,
		cBuffer:        C.calloc(ScanSessionCBufferSize, C.sizeof_char),
		//init C.struct_lis_scan_parameters to it's default which is supposedly zeroed memory
	}
	return &session, nil
}

//ScanSession is just a scan session
type ScanSession struct {
	lisScanSession *C.struct_lis_scan_session
	//ecError        *CErrorProxy
	cBuffer unsafe.Pointer
}

//EndOfFeed indicates that there are no more to read from scanner
func (s *ScanSession) EndOfFeed() bool {
	return C.lis_scan_session_end_of_feed_proxy(s.lisScanSession) != 0
}

//EndOfPage indicates that the current page is over
func (s *ScanSession) EndOfPage() bool {
	return C.lis_scan_session_end_of_page_proxy(s.lisScanSession) != 0
}

//ScanRead reads data from scanner
func (s *ScanSession) ScanRead() ([]byte, uint64, error) {
	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)
	//var err error
	var arrlen C.size_t = ScanSessionCBufferSize
	C.lis_scan_session_scan_read_proxy(s.lisScanSession, s.cBuffer, &arrlen, errProxy.GetProxy())
	if errProxy.ErrNum() != LisOk {
		return nil, 0, errors.New(errProxy.Error())
	}

	//Very very strong magic here
	data := (*[maxSliceLen]byte)(s.cBuffer)[:arrlen:arrlen]
	return data, uint64(arrlen), nil
}

func (s *ScanSession) Cancel() {
	C.lis_scan_session_cancel_proxy(s.lisScanSession)
}

//GetScanParameters returns scanning session parameters
func (s *ScanSession) GetScanParameters() (*ScanParameters, error) {
	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)
	var params ScanParameters
	C.lis_scan_session_get_scan_parameters_proxy(s.lisScanSession, &params.lisScanParameters, errProxy.GetProxy())
	var err error
	if errProxy.ErrNum() != LisOk {
		err = errors.New(errProxy.Error())
	}
	return &params, err
}

//Close frees all allocated resources
func (s *ScanSession) Close() {
	C.free(s.cBuffer)
}

//Width of the image in pixels. This value is guaranteed to be true when scanning
func (sp *ScanParameters) Width() int {
	return int(sp.lisScanParameters.width)
}

//Height of the image in pixels. warning This value is *not* guaranteed to be true when scanning.
func (sp *ScanParameters) Height() int {
	return int(sp.lisScanParameters.height)
}

//ImageFormat is image format. This value is guaranteed to be true when scanning.
func (sp *ScanParameters) ImageFormat() uint32 {
	return sp.lisScanParameters.format
}

//ImageFormatStr returns name of the image format
func (sp *ScanParameters) ImageFormatStr() string {
	return lisImageFormatNames[sp.ImageFormat()]
}


//ImageSize is estimated image size in bytes. Can be used to pre-allocate memory.
//This value is *not* guaranteed to be true when scanning.
func (sp *ScanParameters) ImageSize() uint {
	return uint(sp.lisScanParameters.image_size)
}

func (sp *ScanParameters) String() string {
	return fmt.Sprintf("WxH: %dx%d\n"+
		"Size: %d bytes\n"+
		"Format: %d (%s)\n",
		sp.Width(), sp.Height(), sp.ImageSize(), sp.ImageFormat(), lisImageFormatNames[sp.ImageFormat()])
}

//ListSources gets all scan sources (flatbed, auto-feeder)
/*
func (d *LisDeviceDescriptor) ListSources() []*PaperSource {
	sources := C.go_list_sources(d.LisAPI.lisAPI, d.lisDevice)
	len := C.go_lis_array_length(unsafe.Pointer(sources))
	const maxLen = 1 << 30
	//Apply strong magic to get GO slice backed by C null-terminated array
	slice := (*[maxLen]*C.struct_lis_item)(unsafe.Pointer(sources))[:len:len]
	fmt.Printf("SOURCES: %+v\n", slice)
	var res []*PaperSource
	for _, s := range slice {
		res = append(res,
			&PaperSource{
				s,
				C.GoString(s.name),
			})
	}
	return res
}*/

package lisgo

/*
#include <libinsane/capi.h>
#include <lislib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

//lis_value_type enum
const (
	LisTypeBool = iota
	LisTypeInteger
	LisTypeDouble
	LisTypeString
	LisTypeImageFormat
)

//lis_constraint_type enum
const (
	LisConstraintNone = iota
	LisConstraintRange
	LisConstraintList
)

//capabilitites bit mask flags
const (
	LisCapEmulated  = 1 << 0 // emulated option (either by LibInsane or backend)
	LisCapAutomatic = 1 << 1 // can be set automatically by the driver
	LisCapHwSelect  = 1 << 2 // can be set by a hardware switch
	LisCapSwSelect  = 1 << 3 //  read/write ; can be set by software
	LisCapInactive  = 1 << 4 //requires another option to be set to a specific value to become active
)

/*
 * Image format.
 *
 * Used to defined the type of content that will be returned by \ref lis_scan_session.scan_read() . */
const (

	/* Raw image, 24bits per pixel.
	*
	* For each pixel:
	* - 8bits for red,
	* - 8bits for green,
	* - 8bits for blue.
	*
	* No header, just pixels. */
	LisImgFormatRawRGB24 = iota
	LisImgFormatGrayScale8
	LisImgFormatBW1
	LisImgFormatBmp
	LisImgFormatCiff
	LisImgFormatExif
	LisImgFormatFlashPix
	LisImgFormatGif
	LisImgFormatJpeg
	LisImgFormatPng
	LisImgFormatIco
	LisImgFormatJpeg2k
	LisImgFormatJpeg2kx
	LisImgFormatmMemoryBmp // Windows BMP without header
	LisImgFormatPhotoCD
	LisImgFormatPict
	LisImgFormatTiff
)

//enum lis_unit
const (
	LisUnitNone = iota
	LisUnitPixel
	LisUnitBit
	LisUnitMM
	LisUnitDPI
	LisUnitPercent
	LisUnitMicrosecond
)

var (
	lisCapNames = map[int]string{
		LisCapEmulated:  "LisCapEmulated",
		LisCapAutomatic: "LisCapAutomatic",
		LisCapHwSelect:  "LisCapHwSelect",
		LisCapSwSelect:  "LisCapSwSelect",
		LisCapInactive:  "LisCapInactive"}

	lisTypeNames = map[uint32]string{
		LisTypeBool:        "Bool",
		LisTypeInteger:     "Integer",
		LisTypeDouble:      "Double",
		LisTypeString:      "String",
		LisTypeImageFormat: "ImageFormat"}

	lisConstraintNames = map[int]string{
		LisConstraintNone:  "None",
		LisConstraintRange: "Range",
		LisConstraintList:  "List"}

	lisImageFormatNames = map[uint32]string{
		LisImgFormatRawRGB24:   "Raw RGB24",
		LisImgFormatGrayScale8: "Grayscale 8",
		LisImgFormatBW1:        "BW1",
		LisImgFormatBmp:        "BMP",
		LisImgFormatCiff:       "CIFF",
		LisImgFormatExif:       "EXIF",
		LisImgFormatFlashPix:   "FlashPix",
		LisImgFormatGif:        "GIF",
		LisImgFormatJpeg:       "JPEG",
		LisImgFormatPng:        "PNG",
		LisImgFormatIco:        "ICO",
		LisImgFormatJpeg2k:     "JPEG2K",
		LisImgFormatJpeg2kx:    "JPEG2KX",
		LisImgFormatmMemoryBmp: "Memory BMP", // Windows BMP without header
		LisImgFormatPhotoCD:    "PhotoCD",
		LisImgFormatPict:       "Pict",
		LisImgFormatTiff:       "TIFF",
	}

	lisUnitNames = map[uint32]string{
		LisUnitNone:        "None",
		LisUnitPixel:       "Pixel",
		LisUnitBit:         "Bit",
		LisUnitMM:          "MM",
		LisUnitDPI:         "DPI",
		LisUnitPercent:     "Percent",
		LisUnitMicrosecond: "Microsecond",
	}
)

type (
	//LisValue is value of option
	LisValue struct {
		ValType     uint32
		BoolValue   bool
		IntValue    int
		DoubleValue float64
		StringValue string
		ImgFormat   C.enum_lis_img_format
	}

	//ValueRange define constraints applied to value
	ValueRange struct {
		MinValue *LisValue
		MaxValue *LisValue
		Interval *LisValue
	}

	//ValueList is an array of values
	ValueList []*LisValue

	//OptionDescriptor is a definition of an scan option
	OptionDescriptor struct {
		//Name option name / identifier (ex: "source", "resolution", etc).
		Name string
		//Title Human readable title (usually in English).
		Title string
		//Desc Human readable description (usually in English).
		Desc string
		/* Capabilities is the option capabilities.
		*
		* Bit fields:
		* -  LIS_CAP_EMULATED
		* -  LIS_CAP_AUTOMATIC
		* -  LIS_CAP_HW_SELECT
		* -  LIS_CAP_SW_SELECT
		* -  LIS_CAP_INACTIVE */
		Capabilities int
		// Type of this option.
		ValueType C.enum_lis_value_type
		// Unit of this value. Only useful for integers and float.
		ValueUnit  C.enum_lis_unit
		Constraint *OptionConstraint
		optStruct  *C.struct_lis_option_descriptor
	}

	//OptionConstraint describe restrictions defining the possible values for this option.
	OptionConstraint struct {
		/*
			LIS_CONSTRAINT_NONE = 0,  No constraint
			LIS_CONSTRAINT_RANGE = 1, Range of values (integers only)
			LIS_CONSTRAINT_LIST = 2, The structure 'possible' contains a list of values
		*/
		ConstraintType int
		//If LIS_CONSTRAINT_RANGE.
		PossibleRange *ValueRange
		//If  LIS_CONSTRAINT_LIST.
		PossibleList ValueList
	}
)

//GetValue obtains value of an option
func (o *OptionDescriptor) GetValue() (*LisValue, error) {
	if !o.IsReadable() {
		return nil, errors.New("сannot read the option")
	}

	errProxy := getErrorProxy()
	defer releaseErrorProxy(errProxy)
	var err error
	val := C.lis_option_descriptor_get_value_proxy(o.optStruct, errProxy.GetProxy())
	if errProxy.ErrNum() != LisOk {
		err = errors.New(errProxy.Error())
		return nil, err
	}
	return NewValue(val, o.ValueType), nil

}

//Print option using fmt
func (o *OptionDescriptor) String() string {
	var valStr = ""

	if o.IsReadable() {
		val, err := o.GetValue()
		if err != nil {
			valStr = err.Error() + "\n"
		} else {
			valStr = fmt.Sprintf("Value: %v\n", val)
		}
	}
	return fmt.Sprintf(
		"%s (%s;%s)\n"+
			"Caps: %v %s\n"+
			"Type: %v (%v)\n"+
			"Units: %v (%s)\n"+
			"Constraint: %v\n",
		o.Name, o.Title, o.Desc,
		o.Capabilities, o.formatCaps(),
		o.ValueType, o.formatType(),
		o.ValueUnit, lisUnitNames[o.ValueUnit],
		o.formatConstraint()) + valStr
}

func (o *OptionDescriptor) bitTest(mask int) bool {
	return (o.Capabilities & mask) == mask
}

//IsReadable indicates that option can be read
func (o *OptionDescriptor) IsReadable() bool {
	return !o.bitTest(LisCapInactive)
}

//IsWritable indicates that option can be written
func (o *OptionDescriptor) IsWritable() bool {
	return o.bitTest(LisCapSwSelect)
}

func (o *OptionDescriptor) formatCaps() string {
	caps := ""
	for k, v := range lisCapNames {
		if o.bitTest(k) {
			caps += v + ","
		}
	}
	if caps == "" {
		return "[]"
	}

	return fmt.Sprintf("[%s]", caps[:len(caps)-1])
}

func (o *OptionDescriptor) formatType() string {
	return lisTypeNames[o.ValueType]
}

func (o *OptionDescriptor) formatConstraint() string {
	//cons := lisConstraintNames[o.Constraint.ConstraintType]
	cons := ""
	if o.Constraint.ConstraintType == LisConstraintList {
		for _, v := range o.Constraint.PossibleList {
			cons += v.String() + ","
		}
		if len(cons) > 0 {
			cons = cons[:len(cons)-1]
		}
	}
	if o.Constraint.ConstraintType == LisConstraintRange {
		cons = fmt.Sprintf("min: %v, max: %v, interval: %v",
			o.Constraint.PossibleRange.MinValue,
			o.Constraint.PossibleRange.MaxValue,
			o.Constraint.PossibleRange.Interval)
	}
	name := lisConstraintNames[o.Constraint.ConstraintType]
	if cons == "" {
		return name
	}

	return fmt.Sprintf("%s (%s)", name, cons)

}

//Print prints LisValue depending on type
func (v *LisValue) String() string {
	switch v.ValType {
	case LisTypeBool:
		return fmt.Sprintf("%v", v.BoolValue)
	case LisTypeInteger:
		return fmt.Sprintf("%v", v.IntValue)
	case LisTypeDouble:
		return fmt.Sprintf("%v", v.DoubleValue)
	case LisTypeString:
		return fmt.Sprintf("%v", v.StringValue)
	case LisTypeImageFormat:
		return fmt.Sprintf("%v", v.ImgFormat)
	default:
		panic("Unknown value type")
	}
}

//NewValue constructs GO LisValue struct from lis_value C-struct
func NewValue(val *C.union_lis_value, typ C.enum_lis_value_type) *LisValue {
	var res LisValue
	res.ValType = typ
	switch res.ValType {
	case LisTypeBool:
		res.BoolValue = *((*int)(unsafe.Pointer(val))) != 0
	case LisTypeInteger:
		res.IntValue = int(*((*C.int)(unsafe.Pointer(val))))
	case LisTypeDouble:
		res.DoubleValue = *((*float64)(unsafe.Pointer(val)))
	case LisTypeString:
		res.StringValue = C.GoString(*(**C.char)(unsafe.Pointer(val)))
	case LisTypeImageFormat:
		res.ImgFormat = *(*C.enum_lis_img_format)(unsafe.Pointer(val))
	default:
		panic("Unknown value type")
	}
	return &res
}

//NewConstraint costructs Go ValueConstraint struct from C-structs
func NewConstraint(valType C.enum_lis_value_type, conType C.int, conPossible unsafe.Pointer) *OptionConstraint {
	con := OptionConstraint{
		ConstraintType: int(conType),
	}

	//conRange *C.struct_lis_value_range, conList *C.struct_lis_value_list

	if con.ConstraintType == LisConstraintList {
		conList := (*C.struct_lis_value_list)(conPossible)
		//fmt.Printf("GO array address: %p\n", conList.values)
		//C.lis_value_array_print(valType, conList.values, conList.nb_values)
		con.PossibleList = []*LisValue{}

		slice := carrToSlice(conList.values, int(conList.nb_values))
		for i := range slice {
			//C.lis_value_print(valType, &v)
			c := NewValue(&slice[i], valType)
			//fmt.Printf("GO element %p\n", &slice[i])
			//fmt.Println(t)
			con.PossibleList = append(con.PossibleList, c)

		}
		//fmt.Printf("%+v", con.PossibleList)
	}

	if con.ConstraintType == LisConstraintRange {
		conRange := (*C.struct_lis_value_range)(conPossible)

		var c ValueRange
		c.MinValue = NewValue(&conRange.min, valType)
		c.MaxValue = NewValue(&conRange.max, valType)
		c.Interval = NewValue(&conRange.interval, valType)
		con.PossibleRange = &c
	}
	return &con
}

func carrToSlice(arr *C.union_lis_value, count int) []C.union_lis_value {

	//Apply strong magic to get GO slice backed by C null-terminated array
	slice := (*[maxSliceLen]C.union_lis_value)(unsafe.Pointer(arr))[:count:count]
	return slice
}

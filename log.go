package lisgo

/*
#include <libinsane/log.h>
#include <lislib.h>

*/
import "C"
import (
	"github.com/apex/log"
)

//Log levels from libinsane
const (
	LisLogLvlMin     uint32 = 0
	LisLogLvlDebug   uint32 = LisLogLvlMin
	LisLogLvlInfo    uint32 = 1
	LisLogLvlWarning uint32 = 2
	LisLogLvlError   uint32 = 3
	LisLogLvlMax     uint32 = LisLogLvlError
)

//Lis errors
const (
	LisOk uint32 = 0
)

var (
	logLevel uint32 = LisLogLvlError
)

//SetLogLevel sets min log level
func SetLogLevel(lvl uint32) {
	logLevel = lvl
}

//export logProxy
func logProxy(lvl C.enum_lis_log_level, msg *C.char) {
	if lvl >= logLevel {
		if lvl >= LisLogLvlError {
			log.WithField("error", C.GoString(msg)).Error("libinsane error")
		} else {
			log.WithField("message", C.GoString(msg)).Debug("libinsane log message")
		}
		//C.go_printf(msg)
	}
}

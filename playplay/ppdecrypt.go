package playplay

/*
#cgo CXXFLAGS: -std=c++17
#include "ppdecrypt.h"
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"
)

func PlayPlayDecrypt(keyBasis [16]byte, fileID [20]byte) [16]byte {
	var buf [16]byte

	C.ppdecrypt(
		(*C.uint8_t)(unsafe.Pointer(&keyBasis[0])),
		(*C.uint8_t)(unsafe.Pointer(&fileID[0])),
		(*C.uint8_t)(unsafe.Pointer(&buf[0])),
	)
	return buf
}

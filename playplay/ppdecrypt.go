package playplay

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#include "src/ppdecrypt.h"
#include <stdlib.h>

// 导出函数声明
extern void ppdecrypt(const uint8_t key_basis[16], const uint8_t file_id[20], uint8_t dst[16]);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func Test() {
	// 定义 Go 的 key_basis, file_id 和 dst
	keyBasis := [16]uint8{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}
	fileId := [20]uint8{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13}
	var dst [16]uint8

	// 调用 C 函数
	C.ppdecrypt(
		(*C.uint8_t)(unsafe.Pointer(&keyBasis[0])),
		(*C.uint8_t)(unsafe.Pointer(&fileId[0])),
		(*C.uint8_t)(unsafe.Pointer(&dst[0])),
	)

	// 打印结果
	fmt.Printf("Decrypted Data: %v\n", dst)
}

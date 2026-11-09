//go:build windows

package window

import (
	"log"
	"syscall"
	"unsafe"

	"github.com/egoist/mygo"
)

var procDwmSetWindowAttribute = syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")

func Show(win *mygo.Window) {
	var cloaked int32
	result, _, _ := procDwmSetWindowAttribute.Call(win.NativeHandle(), 13, uintptr(unsafe.Pointer(&cloaked)), unsafe.Sizeof(cloaked))
	if int32(result) < 0 {
		log.Printf("解除窗口合成隐藏失败：HRESULT %#x", uint32(result))
	}
	win.Show()
}

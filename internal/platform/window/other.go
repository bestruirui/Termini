//go:build !windows

package window

import "github.com/egoist/mygo"

func Show(win *mygo.Window) {
	win.Show()
}

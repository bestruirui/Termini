//go:build !windows

package explorer

import "errors"

// SetContextMenu 在其他平台报告不支持 Windows 文件管理器菜单。
func SetContextMenu(bool) error {
	return errors.New("仅 Windows 支持文件管理器右键菜单注册")
}

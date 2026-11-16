//go:build !windows

package explorer

import "errors"

// RunAsAdministrator 在其他平台报告不支持 Windows 提权启动。
func RunAsAdministrator(string) error {
	return errors.New("仅 Windows 支持启动 (管理员)")
}

// SetContextMenu 在其他平台报告不支持 Windows 文件管理器菜单。
func SetContextMenu(bool, bool) error {
	return errors.New("仅 Windows 支持文件管理器右键菜单注册")
}

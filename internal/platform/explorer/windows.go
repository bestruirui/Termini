//go:build windows

package explorer

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	advapi32           = syscall.NewLazyDLL("advapi32.dll") // Windows 注册表接口。
	procRegCreateKeyEx = advapi32.NewProc("RegCreateKeyExW") // 打开或创建注册项。
	procRegSetValueEx  = advapi32.NewProc("RegSetValueExW")  // 写入注册项的字符串值。
	procRegDeleteTree  = advapi32.NewProc("RegDeleteTreeW")  // 删除菜单注册项及其子项。
)

// RunAsAdministrator 请求 Windows 提权，并在指定目录重新启动当前程序。
func RunAsAdministrator(workingDirectory string) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	file, err := syscall.UTF16PtrFromString(executable)
	if err != nil {
		return err
	}
	arguments, err := syscall.UTF16PtrFromString("--working-directory " + syscall.EscapeArg(workingDirectory))
	if err != nil {
		return err
	}
	verb, _ := syscall.UTF16PtrFromString("runas")
	// 不传递提权参数，避免新进程重复请求提权；用户取消时直接结束启动。
	if err := windows.ShellExecute(0, verb, file, arguments, nil, windows.SW_SHOWNORMAL); err != nil && !errors.Is(err, windows.ERROR_CANCELLED) {
		return fmt.Errorf("启动 (管理员) 失败：%w", err)
	}
	return nil
}

// SetContextMenu 为当前用户注册或移除指定权限的文件夹和空白处 Termini 右键菜单。
func SetContextMenu(enabled, administrator bool) error {
	var executable string // 开启菜单时使用当前程序的绝对路径。
	if enabled {
		var err error
		executable, err = os.Executable()
		if err != nil {
			return err
		}
	}
	// 文件夹传入选中项路径，空白处传入当前目录；各菜单只修改本应用的注册项。
	for _, entry := range []struct {
		path          string // 当前用户下的菜单注册项路径。
		target        string // 资源管理器传入的目录占位符。
		administrator bool   // 是否通过 Windows 提权启动。
	}{
		{`Software\Classes\Directory\shell\Termini`, "%1", false},
		{`Software\Classes\Directory\shell\TerminiAdmin`, "%1", true},
		{`Software\Classes\Directory\Background\shell\Termini`, "%V", false},
		{`Software\Classes\Directory\Background\shell\TerminiAdmin`, "%V", true},
	} {
		if entry.administrator != administrator {
			continue
		}
		if !enabled {
			path, err := syscall.UTF16PtrFromString(entry.path)
			if err != nil {
				return err
			}
			result, _, _ := procRegDeleteTree.Call(uintptr(syscall.HKEY_CURRENT_USER), uintptr(unsafe.Pointer(path)))
			if result != 0 && !errors.Is(syscall.Errno(result), os.ErrNotExist) {
				return fmt.Errorf("删除右键菜单失败：%w", syscall.Errno(result))
			}
			continue
		}
		// 目录末尾补上 \.，避免盘符根目录的尾部反斜杠影响命令行引号。
		label := "使用 Termini 打开"
		command := fmt.Sprintf(`"%s" --working-directory "%s\."`, executable, entry.target)
		if entry.administrator {
			label = "使用 Termini 打开 (管理员)"
			command = fmt.Sprintf(`"%s" --administrator --working-directory "%s\."`, executable, entry.target)
		}
		for _, value := range []struct {
			path string // 要写入的注册项路径。
			name string // 值名称，空值写入默认项。
			data string // 菜单名称、图标或启动命令。
		}{
			{entry.path, "", label},
			{entry.path, "Icon", fmt.Sprintf(`"%s",0`, executable)},
			{entry.path, "HasLUAShield", ""},
			{entry.path + `\command`, "", command},
		} {
			if value.name == "HasLUAShield" && !entry.administrator {
				continue
			}
			path, err := syscall.UTF16PtrFromString(value.path)
			if err != nil {
				return err
			}
			name, err := syscall.UTF16PtrFromString(value.name)
			if err != nil {
				return err
			}
			data, err := syscall.UTF16FromString(value.data)
			if err != nil {
				return err
			}
			var key syscall.Handle // 写入字符串期间持有的注册表句柄。
			result, _, _ := procRegCreateKeyEx.Call(
				uintptr(syscall.HKEY_CURRENT_USER), uintptr(unsafe.Pointer(path)),
				0, 0, 0, syscall.KEY_SET_VALUE, 0, uintptr(unsafe.Pointer(&key)), 0,
			)
			if result != 0 {
				return fmt.Errorf("创建右键菜单注册项失败：%w", syscall.Errno(result))
			}
			result, _, _ = procRegSetValueEx.Call(
				uintptr(key), uintptr(unsafe.Pointer(name)), 0, syscall.REG_SZ,
				uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)*2),
			)
			syscall.RegCloseKey(key)
			if result != 0 {
				return fmt.Errorf("写入右键菜单失败：%w", syscall.Errno(result))
			}
		}
	}
	return nil
}

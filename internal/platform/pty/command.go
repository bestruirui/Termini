package pty

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
)

// Shells 按默认启动优先级列出本机可用的 shell 可执行文件。
func Shells() []string {
	names := []string{os.Getenv("SHELL"), "/bin/sh", "bash", "zsh", "fish"}
	if runtime.GOOS == "darwin" {
		names = []string{os.Getenv("SHELL"), "/bin/zsh", "/bin/sh", "bash", "fish"}
	} else if runtime.GOOS == "windows" {
		names = []string{"pwsh.exe", "powershell.exe", os.Getenv("ComSpec"), "cmd.exe", "wsl.exe"}
	}
	var paths []string
	for _, name := range names {
		if name == "" {
			continue
		}
		path, err := exec.LookPath(name)
		if err == nil && !slices.Contains(paths, path) {
			paths = append(paths, path)
		}
	}
	return paths
}

// Command 解析启动程序并检查原始参数；自动选择且没有参数时保留终端原有默认行为。
func Command(shell, arguments string) ([]string, error) {
	if strings.IndexByte(arguments, 0) >= 0 {
		return nil, errors.New("启动参数不能包含空字符")
	}
	if shell == "" {
		if arguments == "" {
			return nil, nil
		}
		shells := Shells()
		if len(shells) == 0 {
			return nil, errors.New("没有找到可用的 shell")
		}
		shell = shells[0]
	}
	path, err := exec.LookPath(shell)
	if err != nil {
		return nil, fmt.Errorf("找不到启动程序 %q：%w", shell, err)
	}
	return []string{path}, nil
}

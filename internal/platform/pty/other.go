//go:build !windows

package pty

import "github.com/egoist/mygo/plugins/terminal"

// NewTerminal 在其他平台使用 MyGo 的系统 PTY。
// opts.Command 必须来自 Command；非空参数行交由系统 shell 解析。
func NewTerminal(opts terminal.Options, arguments string) (*terminal.Terminal, error) {
	if arguments != "" {
		opts.Command = []string{"/bin/sh", "-c", "exec \"$0\" " + arguments, opts.Command[0]}
	}
	return terminal.New(opts)
}

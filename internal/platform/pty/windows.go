//go:build windows

package pty

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// ptyConn 将随应用分发的 ConPTY 接入终端的读写和尺寸通知。
type ptyConn struct {
	mu      sync.Mutex        // 保护控制台句柄和尺寸。
	console uintptr           // 控制台句柄，关闭后为零。
	cols    int               // 当前列数。
	rows    int               // 当前行数。
	in      *os.File          // 写入控制台的输入管道。
	out     *os.File          // 读取控制台输出的管道。
	resize  *syscall.LazyProc // 调整控制台尺寸。
	clear   *syscall.LazyProc // 清除控制台已保存的画面。
	close   *syscall.LazyProc // 关闭控制台及其附属进程。
}

// startupInfoEx 为进程启动信息附加控制台属性列表。
type startupInfoEx struct {
	syscall.StartupInfo         // Windows 进程启动信息。
	attributes         uintptr // 进程属性列表的地址。
}

var (
	kernel32                              = syscall.NewLazyDLL("kernel32.dll")                // Windows 内核接口。
	procInitializeProcThreadAttributeList = kernel32.NewProc("InitializeProcThreadAttributeList") // 分配进程属性列表。
	procUpdateProcThreadAttribute         = kernel32.NewProc("UpdateProcThreadAttribute")         // 设置进程的控制台属性。
	procDeleteProcThreadAttributeList     = kernel32.NewProc("DeleteProcThreadAttributeList")     // 释放进程属性列表。
)

// NewTerminal 在随应用分发的 ConPTY 中启动用户的 shell，并连接终端视图。
// arguments 是未经拆分或转义的单行启动参数。
func NewTerminal(opts terminal.Options, arguments string) (*terminal.Terminal, error) {
	dir, err := mygo.App.Path(mygo.PathResources)
	if err != nil {
		return nil, err
	}
	// 开发和打包后资源位于顶层；从工程目录直接运行时读取平台资源。
	if _, err := os.Stat(filepath.Join(dir, "conpty.dll")); errors.Is(err, os.ErrNotExist) {
		dir, err = filepath.Abs(filepath.Join("resources", "windows-"+runtime.GOARCH))
		if err != nil {
			return nil, err
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "OpenConsole.exe")); err != nil {
		return nil, fmt.Errorf("pty: OpenConsole.exe: %w", err)
	}
	dll := syscall.NewLazyDLL(filepath.Join(dir, "conpty.dll"))
	create := dll.NewProc("CreatePseudoConsole")
	p := &ptyConn{
		cols:   80,
		rows:   24,
		resize: dll.NewProc("ResizePseudoConsole"),
		clear:  dll.NewProc("ClearPseudoConsole"),
		close:  dll.NewProc("ClosePseudoConsole"),
	}
	for _, proc := range []*syscall.LazyProc{create, p.resize, p.clear, p.close} {
		if err := proc.Find(); err != nil {
			return nil, fmt.Errorf("pty: %w", err)
		}
	}
	var shell string
	if len(opts.Command) > 0 {
		shell, err = exec.LookPath(opts.Command[0])
		if err != nil {
			return nil, fmt.Errorf("pty: %w", err)
		}
	} else {
		for _, name := range []string{"pwsh.exe", "powershell.exe", os.Getenv("ComSpec"), "cmd.exe"} {
			if name == "" {
				continue
			}
			if shell, err = exec.LookPath(name); err == nil {
				break
			}
		}
	}
	if shell == "" {
		return nil, errors.New("pty: no shell found")
	}
	// 仅转义可执行文件路径，启动参数原样追加到命令行。
	commandLine := syscall.EscapeArg(shell)
	if arguments != "" {
		commandLine += " " + arguments
	}
	command, err := syscall.UTF16PtrFromString(commandLine)
	if err != nil {
		return nil, err
	}
	workingDirectory := opts.Dir // 菜单指定目录优先，未指定时沿用用户主目录。
	if workingDirectory == "" {
		workingDirectory, err = os.UserHomeDir()
		if err != nil {
			return nil, err
		}
	}
	cwd, err := syscall.UTF16PtrFromString(workingDirectory)
	if err != nil {
		return nil, err
	}
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.EqualFold(key, "TERM") && !strings.EqualFold(key, "COLORTERM") {
			env = append(env, entry)
		}
	}
	env = append(env, "TERM=xterm-256color", "COLORTERM=truecolor")
	block := utf16.Encode([]rune(strings.Join(env, "\x00") + "\x00\x00"))
	var inR, inW, outR, outW syscall.Handle
	if err := syscall.CreatePipe(&inR, &inW, nil, 0); err != nil {
		return nil, fmt.Errorf("pty: %w", err)
	}
	if err := syscall.CreatePipe(&outR, &outW, nil, 0); err != nil {
		syscall.CloseHandle(inR)
		syscall.CloseHandle(inW)
		return nil, fmt.Errorf("pty: %w", err)
	}
	// 继承终端光标的位置，由终端通过 DSR 响应控制台的位置查询。
	r, _, _ := create.Call(uintptr(80)|uintptr(24)<<16, uintptr(inR), uintptr(outW), 1, uintptr(unsafe.Pointer(&p.console)))
	syscall.CloseHandle(inR)
	syscall.CloseHandle(outW)
	if r != 0 {
		syscall.CloseHandle(inW)
		syscall.CloseHandle(outR)
		return nil, fmt.Errorf("pty: CreatePseudoConsole: %w", syscall.Errno(r&0xFFFF))
	}
	p.in = os.NewFile(uintptr(inW), "conpty-in")
	p.out = os.NewFile(uintptr(outR), "conpty-out")
	connected := false // 接入成功后由终端管理管道的生命周期。
	defer func() {
		if !connected {
			// 启动失败时继续排空输出，避免关闭控制台等待无人读取的管道。
			go io.Copy(io.Discard, p.out)
			p.Close()
		}
	}()
	var size uintptr
	procInitializeProcThreadAttributeList.Call(0, 1, 0, uintptr(unsafe.Pointer(&size)))
	attributes := make([]byte, size)
	if r, _, err := procInitializeProcThreadAttributeList.Call(uintptr(unsafe.Pointer(&attributes[0])), 1, 0, uintptr(unsafe.Pointer(&size))); r == 0 {
		return nil, fmt.Errorf("pty: InitializeProcThreadAttributeList: %w", err)
	}
	defer func() { procDeleteProcThreadAttributeList.Call(uintptr(unsafe.Pointer(&attributes[0]))) }()
	if r, _, err := procUpdateProcThreadAttribute.Call(uintptr(unsafe.Pointer(&attributes[0])), 0, 0x00020016, p.console, unsafe.Sizeof(p.console), 0, 0); r == 0 {
		return nil, fmt.Errorf("pty: UpdateProcThreadAttribute: %w", err)
	}
	startup := startupInfoEx{attributes: uintptr(unsafe.Pointer(&attributes[0]))}
	startup.Title, _ = syscall.UTF16PtrFromString("Termini")
	startup.Cb = uint32(unsafe.Sizeof(startup))
	startup.Flags = syscall.STARTF_USESTDHANDLES
	startup.StdInput, startup.StdOutput, startup.StdErr = syscall.InvalidHandle, syscall.InvalidHandle, syscall.InvalidHandle
	var process syscall.ProcessInformation
	if err := syscall.CreateProcess(nil, command, nil, nil, false, 0x00080000|0x00000400, &block[0], cwd, &startup.StartupInfo, &process); err != nil {
		return nil, fmt.Errorf("pty: starting %s: %w", shell, err)
	}
	syscall.CloseHandle(process.Thread)
	go func() {
		syscall.WaitForSingleObject(process.Process, syscall.INFINITE)
		syscall.CloseHandle(process.Process)
		p.Close()
	}()
	opts.Conn = p
	var term *terminal.Terminal
	// 临时适配 MyGo v0.7.4：缺少 ConPTY 的 Win32 输入协议支持，需直接编码 Shift+Enter 和 Esc。
	// 升级到包含上游修复的 MyGo 版本后，删除此 OnKey 适配及仅供它使用的 ui 导入。
	onKey := opts.OnKey // 保留调用方已有的按键处理及其优先级。
	opts.OnKey = func(mods ui.Modifiers, key ui.Key) bool {
		if onKey != nil && onKey(mods, key) {
			return true
		}
		switch {
		case mods == ui.Shift && key == ui.KeyEnter:
			// 成对发送 Win32 回车按下、释放记录：虚拟键码和 CR 均为 13，Shift 状态位为 16。
			term.Send([]byte("\x1b[13;0;13;1;16;1_\x1b[13;0;13;0;16;1_"))
		case mods == 0 && key == ui.KeyEscape:
			// 使用 VK_ESCAPE 的按下、释放记录，避免 Kitty 的 CSI-u 编码被 ConPTY 当作文本。
			term.Send([]byte("\x1b[27;1;27;1;0;1_\x1b[27;1;27;0;0;1_"))
		default:
			return false
		}
		return true
	}
	opts.OnClear = func(action terminal.Clear) bool {
		if action != terminal.ClearScreen {
			return false
		}
		if term.ClearScreen() {
			p.mu.Lock()
			if p.console != 0 {
				p.clear.Call(p.console)
			}
			p.mu.Unlock()
		}
		return true
	}
	term, err = terminal.New(opts)
	if err != nil {
		return nil, err
	}
	connected = true
	return term, nil
}

// Read 读取控制台输出，管道关闭后报告输出结束。
func (p *ptyConn) Read(b []byte) (int, error) {
	n, err := p.out.Read(b)
	if n == 0 && (errors.Is(err, os.ErrClosed) || errors.Is(err, syscall.ERROR_BROKEN_PIPE)) {
		err = io.EOF
	}
	return n, err
}

// Write 将键盘输入和终端查询的响应写入控制台。
func (p *ptyConn) Write(b []byte) (int, error) { return p.in.Write(b) }

// Resize 先调整列数，再调整行数，保持控制台的光标位置。
func (p *ptyConn) Resize(cols, rows int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.console == 0 {
		return os.ErrClosed
	}
	cols, rows = min(max(cols, 1), 0x7FFF), min(max(rows, 1), 0x7FFF)
	if cols != p.cols && rows != p.rows {
		if r, _, _ := p.resize.Call(p.console, uintptr(uint16(cols))|uintptr(uint16(p.rows))<<16); r != 0 {
			return fmt.Errorf("pty: ResizePseudoConsole: %w", syscall.Errno(r&0xFFFF))
		}
		p.cols = cols
	}
	if r, _, _ := p.resize.Call(p.console, uintptr(uint16(cols))|uintptr(uint16(rows))<<16); r != 0 {
		return fmt.Errorf("pty: ResizePseudoConsole: %w", syscall.Errno(r&0xFFFF))
	}
	p.cols, p.rows = cols, rows
	return nil
}

// Close 关闭输入和控制台，排空输出后释放输出管道。
func (p *ptyConn) Close() error {
	p.mu.Lock()
	console := p.console
	p.console = 0
	p.mu.Unlock()
	if console == 0 {
		return nil
	}
	p.in.Close()
	go func() {
		p.close.Call(console)
		p.out.Close()
	}()
	return nil
}

package settingwindow

import (
	"slices"
	"strings"

	"termini/internal/component"
	"termini/internal/platform/pty"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// terminalPage 即时调整终端文字，并编辑下次启动使用的光标、右键行为、程序和参数。
func (w *Window) terminalPage(c *ui.Context) {
	changed := false // 本次界面构建是否收到终端文字设置变更。
	group(c, "文字", func() {
		row(c, "终端字体", func() {
			font := w.preferences.Shell.FontFamily
			if font == "" {
				font = "系统等宽字体"
			}
			if component.SearchSelect(c.Key("font"), &font, w.terminalFontNames, 240) {
				if font == "系统等宽字体" {
					font = ""
				}
				w.preferences.Shell.FontFamily = font
				changed = true
			}
		})
		ui.Divider(c)
		row(c, "字号", func() {
			var sizes []float64
			for size := 8; size <= 48; size++ {
				sizes = append(sizes, float64(size))
			}
			if !slices.Contains(sizes, w.preferences.Shell.FontSize) {
				sizes = append(sizes, w.preferences.Shell.FontSize)
				slices.Sort(sizes)
			}
			changed = component.Select(c.Key("font-size"), &w.preferences.Shell.FontSize, sizes, 90).Changed() || changed
		})
	})
	if changed {
		if w.term != nil {
			w.term.SetFont(terminal.Font{Family: w.preferences.Shell.FontFamily, Size: float32(w.preferences.Shell.FontSize)})
		}
		w.save()
	}
	ui.Column(c).Gap(8).Children(func() {
		group(c, "光标", func() {
			row(c, "光标样式", func() {
				styles := []string{"bar", "block", "underline"}
				labels := []string{"竖线", "块状", "下划线"}
				style := labels[slices.Index(styles, w.preferences.Shell.CursorStyle)]
				if component.Select(c.Key("cursor-style"), &style, labels, 150).Changed() {
					w.preferences.Shell.CursorStyle = styles[slices.Index(labels, style)]
					w.save()
				}
			})
			ui.Divider(c)
			row(c, "光标闪烁", func() {
				if ui.Switch(c.Key("cursor-blink"), &w.preferences.Shell.CursorBlink).Label("光标闪烁").Changed() {
					w.save()
				}
			})
		})
		ui.Text(c, "光标设置在下次启动 Termini 时生效。").FontSize(c.Theme().FontSize-1).TextColor(c.Theme().TextMuted).LineHeight(1.4).Padding(0, 4)
	})
	ui.Column(c).Gap(8).Children(func() {
		group(c, "鼠标", func() {
			row(c, "鼠标右键的作用", func() {
				action := "复制粘贴"
				if w.preferences.Shell.RightClickMenu {
					action = "右键菜单"
				}
				if component.Select(c.Key("right-click"), &action, []string{"复制粘贴", "右键菜单"}, 150).Changed() {
					w.preferences.Shell.RightClickMenu = action == "右键菜单"
					w.save()
				}
			})
		})
		ui.Text(c, "鼠标右键设置在下次启动 Termini 时生效。").FontSize(c.Theme().FontSize-1).TextColor(c.Theme().TextMuted).LineHeight(1.4).Padding(0, 4)
	})
	ui.Column(c).Gap(8).Children(func() {
		group(c, "启动程序", func() {
			row(c, "Shell", func() {
				if component.Select(c.Key("shell"), &w.shellChoice, w.shellNames, 240).Changed() {
					w.terminalNotice = ""
					if w.shellChoice == "自动选择" {
						w.shellDraft = ""
					} else if w.shellChoice != "自定义" {
						w.shellDraft = w.shellChoice
					}
					if w.shellChoice != "自定义" {
						w.saveStartup()
					}
				}
			})
			if w.shellChoice == "自定义" {
				ui.Divider(c)
				row(c, "自定义路径", func() {
					input := ui.TextInput(c.Key("custom-shell"), &w.shellDraft).Width(240).Height(max(float32(26), c.Theme().FontSize+13)).Padding(0, 8)
					input.Placeholder("可执行文件路径或名称").Label("自定义路径")
					if input.Changed() {
						w.terminalNotice = ""
					}
					wasFocused := ui.Local(input, "focused", func() bool { return false }) // 上一帧是否具有键盘焦点。
					if *wasFocused && !input.Focused() {
						w.saveStartup()
					}
					*wasFocused = input.Focused()
				})
			}
			ui.Divider(c)
			row(c, "启动参数", func() {
				input := ui.TextInput(c.Key("arguments"), &w.argumentsDraft).Width(240).Height(max(float32(26), c.Theme().FontSize+13)).Padding(0, 8).Label("启动参数")
				if input.Changed() {
					w.terminalNotice = ""
				}
				wasFocused := ui.Local(input, "focused", func() bool { return false }) // 上一帧是否具有键盘焦点。
				if *wasFocused && !input.Focused() {
					w.saveStartup()
				}
				*wasFocused = input.Focused()
			})
		})
		ui.Text(c, "启动配置在下次启动 Termini 时生效。").FontSize(c.Theme().FontSize-1).TextColor(c.Theme().TextMuted).LineHeight(1.4).Padding(0, 4)
	})
	if w.terminalNotice != "" {
		ui.Text(c, w.terminalNotice).FontSize(c.Theme().FontSize-1).Padding(0, 4).TextColor(c.Theme().Success)
	}
}

// saveStartup 检查并保存启动程序及原始参数，失败时保留草稿并显示错误。
func (w *Window) saveStartup() {
	shell := strings.TrimSpace(w.shellDraft)
	if w.shellChoice == "自定义" && shell == "" {
		w.Error = "请输入启动程序的路径或名称"
		return
	}
	if _, err := pty.Command(shell, w.argumentsDraft); err != nil {
		w.Error = err.Error()
		return
	}
	w.preferences.Shell.Program, w.preferences.Shell.Arguments = shell, w.argumentsDraft
	w.save()
	if w.Error == "" {
		w.terminalNotice = "已保存，下次启动生效"
	}
}

package mainwindow

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

var (
	settingIcon  = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.35" stroke-linejoin="round"><path d="m6.5 2-.4 1.6-1.3.8-1.6-.5L1.7 6.5 3 7.6v.8L1.7 9.5l1.5 2.6 1.6-.5 1.3.8.4 1.6h3l.4-1.6 1.3-.8 1.6.5 1.5-2.6L13 8.4v-.8l1.3-1.1-1.5-2.6-1.6.5-1.3-.8L9.5 2Z"/><circle cx="8" cy="8" r="2"/></svg>`)) // 设置按钮图标。
	minimizeIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.35" stroke-linecap="round"><path d="M3.5 8h9"/></svg>`)) // 最小化按钮图标。
	maximizeIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.35"><rect x="3.5" y="3.5" width="9" height="9" rx="1.5"/></svg>`)) // 最大化按钮图标。
	restoreIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.35" stroke-linecap="round" stroke-linejoin="round"><rect x="3.5" y="5.5" width="7" height="7" rx="1.2"/><path d="M5.5 3.5h5.5a1.5 1.5 0 0 1 1.5 1.5v5.5"/></svg>`)) // 还原窗口按钮图标。
	closeIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.35" stroke-linecap="round"><path d="m3.5 3.5 9 9m0-9-9 9"/></svg>`)) // 关闭窗口按钮图标。
)

// titleBar 绘制可拖动的标题栏、设置按钮和窗口控制按钮，双击空白区域可切换最大化。
// title 用于窗口引用建立前显示该窗口的初始标题。
func titleBar(c *ui.Context, win *mygo.Window, title string, openSettingWindow func()) {
	theme := c.Theme()
	if win != nil {
		title = win.Title()
	}
	bar := ui.Row(c.Key("titlebar")).FillWidth().MinHeight(44).Shrink(0)
	bar.Padding(0, 10).Gap(2).DragWindow()
	bar.TextColor(theme.TextMuted)
	bar.Children(func() {
		ui.Text(c, title).Grow(1).MinWidth(0).MaxLines(1)

		settingButton := ui.ButtonBase(c.Key("settings")).Size(28, 26).Radius(6)
		settingButton.TextColor(theme.TextMuted).Label("设置")
		if settingButton.Pressed() {
			settingButton.Background(theme.Text.Alpha(0.12))
		} else if settingButton.Hovered() {
			settingButton.Background(theme.Text.Alpha(0.05)).TextColor(theme.Text)
		}
		settingButton.Children(func() { ui.Icon(c, settingIcon).Size(16, 16) })
		if settingButton.Clicked() {
			openSettingWindow()
		}

		minimizeButton := ui.ButtonBase(c.Key("minimize")).Size(28, 26).Radius(6)
		minimizeButton.TextColor(theme.TextMuted).Label("最小化")
		if minimizeButton.Pressed() {
			minimizeButton.Background(theme.Text.Alpha(0.12))
		} else if minimizeButton.Hovered() {
			minimizeButton.Background(theme.Text.Alpha(0.05)).TextColor(theme.Text)
		}
		minimizeButton.Children(func() { ui.Icon(c, minimizeIcon).Size(16, 16) })
		if minimizeButton.Clicked() && win != nil {
			win.Minimize()
		}

		maximized := win != nil && win.IsMaximized() // 创建期间可能先绘制首帧，此时按未最大化绘制。
		label, icon := "最大化", maximizeIcon
		if maximized {
			label, icon = "还原", restoreIcon
		}
		maximizeButton := ui.ButtonBase(c.Key("maximize")).Size(28, 26).Radius(6)
		// 标记系统识别的最大化区域，让 Windows 11 可显示贴靠布局。
		maximizeButton.TextColor(theme.TextMuted).Label(label).MaximizeButton()
		if maximizeButton.Pressed() {
			maximizeButton.Background(theme.Text.Alpha(0.12))
		} else if maximizeButton.Hovered() {
			maximizeButton.Background(theme.Text.Alpha(0.05)).TextColor(theme.Text)
		}
		maximizeButton.Children(func() { ui.Icon(c, icon).Size(16, 16) })
		if maximizeButton.Clicked() && win != nil {
			win.ToggleMaximize()
		}

		closeWindowButton(c, win)
	})
}

// closeWindowButton 绘制主窗口与设置窗口共用的关闭按钮，不显示悬浮提示。
func closeWindowButton(c *ui.Context, win *mygo.Window) {
	theme := c.Theme()
	closeButton := ui.ButtonBase(c.Key("close-window")).Size(28, 26).Radius(6)
	closeButton.TextColor(theme.TextMuted).Label("关闭")
	if closeButton.Hovered() || closeButton.Pressed() {
		closeButton.Background(theme.Danger).TextColor(ui.RGB(255, 255, 255))
	}
	closeButton.Children(func() { ui.Icon(c, closeIcon).Size(16, 16) })
	if closeButton.Clicked() && win != nil {
		win.Close()
	}
}

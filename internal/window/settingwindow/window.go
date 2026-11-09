package settingwindow

import (
	"slices"

	"termini/internal/config"
	"termini/internal/platform/pty"
	"termini/internal/platform/window"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/ui"
)

// Window 管理设置窗口、分类页面与尚未保存的启动配置。
type Window struct {
	preferences       *config.Settings   // 与主窗口共享的用户设置。
	term              *terminal.Terminal // 配色与终端文字设置即时应用的终端会话。
	window            *mygo.Window       // 独立设置窗口，关闭后为空。
	page              string             // 当前选中的设置分类。
	Error             string             // 读取、应用或保存设置时的错误。
	themeNames        []string           // 配色选择器中的主题名称。
	appFontNames      []string           // 应用字体选择器中的全部字体名称。
	terminalFontNames []string           // 终端字体选择器中的等宽字体名称。
	shellNames        []string           // 启动程序选择器中的自动项、检测结果和自定义项。
	shellChoice       string             // 启动程序选择器当前选中的项目。
	shellDraft        string             // 尚未保存的启动程序路径或名称。
	argumentsDraft    string             // 尚未保存的单行启动参数，保留原始输入。
	terminalNotice    string             // 启动配置保存成功后的状态提示。
}

var (
	generalIcon    = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.35" stroke-linejoin="round"><path d="m6.5 2-.4 1.6-1.3.8-1.6-.5L1.7 6.5 3 7.6v.8L1.7 9.5l1.5 2.6 1.6-.5 1.3.8.4 1.6h3l.4-1.6 1.3-.8 1.6.5 1.5-2.6L13 8.4v-.8l1.3-1.1-1.5-2.6-1.6.5-1.3-.8L9.5 2Z"/><circle cx="8" cy="8" r="2"/></svg>`))                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      // 通用分类的齿轮图标。
	appearanceIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 56 56"><path fill="currentColor" d="M10.382 24.678c-1.963 1.918-1.85 4.475.25 6.552l4.475 4.52c-2.328 1.484-6.825 3.31-8.971 5.457c-3.356 3.356-3.333 7.876.09 11.323c3.448 3.424 7.968 3.47 11.324.114c2.169-2.169 3.972-6.666 5.456-8.972l4.566 4.498c2.078 2.054 4.612 2.168 6.552.228l2.443-2.511c1.552-1.598 1.803-3.562.685-5.365c1.004-.251 1.94-.822 2.67-1.575l10.616-10.662c2.534-2.51 2.466-5.501-.16-8.127L32.002 1.781C30.357.16 28.121 0 26.591 1.53c-.685.684-1.21 1.666-1.484 2.99c-1.233 5.662-3.722 10.935-6.644 14.657c-.502.662-.844 1.3-1.004 1.963c-1.575-.64-3.22-.297-4.589 1.05Zm26.345 12.374c-.96.936-1.85.867-2.808-.114h-.023L21.614 24.655l.023-.022c-.89-.868-1.302-1.827-.023-3.676c2.648-3.812 5.273-9.04 6.826-14.862c.091-.273.205-.502.41-.707c.366-.388.868-.503 1.37.022l11.095 11.073c-.593 3.812-3.63 7.602-5.958 9.93c-.228.229-.525.548-.16.914c1.142 1.141 7.99-3.767 10.776-6.164l1.62 1.598c1.096 1.096 1.119 2.283.137 3.264ZM13.464 26.847l1.575-1.552c.708-.685 1.644-.708 2.374.045L33.44 41.39c.708.753.73 1.62.023 2.328l-1.507 1.575c-.73.73-1.643.708-2.374-.045l-5.753-5.73c-.799-.822-1.712-.503-2.534.57c-1.666 2.374-4.474 8.356-6.14 10c-1.964 1.917-4.567 1.962-6.507.022s-1.918-4.566.023-6.506c1.643-1.644 7.624-4.475 10.022-6.141c1.05-.822 1.392-1.735.547-2.534l-5.73-5.708c-.73-.707-.753-1.666-.045-2.374m-3.288 21.756c.754.73 1.85.73 2.58 0c.73-.776.708-1.849 0-2.58c-.753-.73-1.872-.753-2.58 0c-.684.777-.73 1.827 0 2.58"/></svg>`)) // 外观分类的画笔图标。
	terminalIcon   = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="m8 9l3 3l-3 3m5 1h3"/></g></svg>`))                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      // 终端分类的控制台图标。
	aboutIcon      = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>`))                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         // 关于分类的信息图标。
)

// Open 打开独立设置窗口，已打开时显示并聚焦已有窗口。
// preferences 必须非空；closeButton 复用应用的关闭按钮。
func (w *Window) Open(
	preferences *config.Settings,
	term *terminal.Terminal,
	closeButton func(*ui.Context, *mygo.Window),
) {
	w.preferences, w.term = preferences, term
	if w.page == "" {
		w.page = "通用"
	}
	if w.window != nil {
		if w.window.IsMinimized() {
			w.window.Restore()
		}
		w.window.Show()
		w.window.Focus()
		return
	}
	w.themeNames = append([]string{"默认"}, terminal.GhosttyThemes()...)
	w.appFontNames = []string{"系统字体"}
	w.terminalFontNames = []string{"系统等宽字体"}
	for _, family := range ui.FontFamilies() {
		w.appFontNames = append(w.appFontNames, family.Name)
		if family.Monospaced {
			w.terminalFontNames = append(w.terminalFontNames, family.Name)
		}
	}
	w.shellNames = append([]string{"自动选择"}, pty.Shells()...)
	w.shellDraft = w.preferences.Shell.Program
	w.shellChoice = w.preferences.Shell.Program
	if w.shellChoice == "" {
		w.shellChoice = "自动选择"
	} else if !slices.Contains(w.shellNames, w.shellChoice) {
		w.shellChoice = "自定义"
	}
	w.shellNames = append(w.shellNames, "自定义")
	w.argumentsDraft = w.preferences.Shell.Arguments
	w.terminalNotice = ""
	w.window = mygo.NewWindow(mygo.WindowOptions{
		Title: "设置 · Termini", Width: 720, Height: 540,
		MinWidth: 660, MinHeight: 480,
		Hidden:            true,
		TitleBarStyle:     mygo.TitleBarHidden,
		HideWindowButtons: true,
		AutoHideMenuBar:   true,
		CornerRadius:      12,
		Content: ui.View(func(c *ui.Context) {
			theme := *c.Theme() // 设置窗口及其菜单统一使用应用文字设置。
			theme.Font, theme.FontSize = w.preferences.Appearance.FontFamily, float32(w.preferences.Appearance.FontSize)
			c.SetTheme(&theme)
			ui.Row(c).Fill().AlignItems(ui.Stretch).Background(c.Theme().Surface).Children(func() {
				w.sidebar(c)
				// 由外层面板裁剪自绘内容，四个圆角外露出窗口底色。
				pane := ui.Box(c).Grow(1).MinWidth(0).MinHeight(0)
				pane.Margin(8).Background(c.Theme().Background).Radius(12).Clip()
				pane.Children(func() { w.view(c, closeButton) })
			})
		}),
	})
	w.window.OnReadyToShow(func() { window.Show(w.window) })
	off := updater.OnChange(w.window.Invalidate) // 窗口关闭时移除更新偏好的重绘监听。
	w.window.OnClosed(func() {
		off()
		w.window = nil
		if w.shellDraft != w.preferences.Shell.Program || w.argumentsDraft != w.preferences.Shell.Arguments {
			w.saveStartup()
		}
	})
}

// sidebar 绘制位于窗口背景层的设置分类。
func (w *Window) sidebar(c *ui.Context) {
	previous := w.page // 切换前的分类，用于保存终端页尚未写入的启动配置。
	if ui.Sidebar(c, &w.page, func() {
		// 图标独立设色，反向排列使后添加的图标仍位于文字前。
		for _, page := range []string{"通用", "外观", "终端", "关于"} {
			item := ui.SidebarItem(c, page, nil, page).TextColor(c.Theme().Text).Reverse()
			if w.page == page {
				item.Background(c.Theme().Background)
			}
			item.Children(func() {
				icon := appearanceIcon
				switch page {
				case "通用":
					icon = generalIcon
				case "终端":
					icon = terminalIcon
				case "关于":
					icon = aboutIcon
				}
				ui.Icon(c, icon).FontSize(16).TextColor(c.Theme().TextMuted)
			})
		}
	}).Width(140).Padding(14, 10, 12, 10).DragWindow().Label("设置分类").Changed() {
		if previous == "终端" && (w.shellDraft != w.preferences.Shell.Program || w.argumentsDraft != w.preferences.Shell.Arguments) {
			w.saveStartup()
		}
		w.window.Invalidate()
	}
}

// view 在圆角内容区顶部绘制分类标题和关闭按钮，下方滚动显示设置分组。
func (w *Window) view(c *ui.Context, closeButton func(*ui.Context, *mygo.Window)) {
	ui.Column(c).Fill().Background(c.Theme().Background).Children(func() {
		ui.Row(c).FillWidth().MinHeight(44).Shrink(0).Padding(0, 10, 0, 24).DragWindow().Children(func() {
			ui.Text(c, w.page).FontSize(c.Theme().FontSize + 2).FontWeight(650).Grow(1).MinWidth(0)
			closeButton(c, w.window)
		})
		ui.Scroll(c.Key("page-"+w.page)).FillWidth().Grow(1).MinHeight(0).Padding(16, 24, 28).Children(func() {
			ui.Column(c).FillWidth().MaxWidth(620).Gap(22).Children(func() {
				switch w.page {
				case "通用":
					w.generalPage(c)
				case "外观":
					w.appearancePage(c)
				case "终端":
					w.terminalPage(c)
				case "关于":
					w.aboutPage(c)
				}
				if w.Error != "" {
					ui.Text(c, w.Error).FontSize(c.Theme().FontSize-1).LineHeight(1.4).Padding(0, 4).TextColor(c.Theme().Danger)
				}
			})
		})
	})
}

// group 用小号标题和浅底圆角面板组织一组相关设置。
func group(c *ui.Context, title string, content func()) {
	ui.Column(c).FillWidth().Gap(8).Children(func() {
		if title != "" {
			ui.Text(c, title).FontSize(c.Theme().FontSize-1).FontWeight(600).TextColor(c.Theme().TextMuted).Padding(0, 4)
		}
		ui.Column(c).FillWidth().Radius(12).Padding(4, 16).Background(c.Theme().Surface).Children(content)
	})
}

// row 将名称放在左侧，操作控件放在右侧。
func row(c *ui.Context, label string, control func()) {
	ui.Row(c.Key("row-"+label)).FillWidth().Padding(11, 0).Gap(16).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).FontSize(c.Theme().FontSize + 0.5).Grow(1).MinWidth(0)
		ui.Row(c).Shrink(0).Gap(8).AlignItems(ui.Center).Children(control)
	})
}

// save 持久化当前设置，并在设置窗口报告保存失败。
func (w *Window) save() {
	if err := config.Save(*w.preferences); err != nil {
		w.Error = "保存设置失败：" + err.Error()
	} else {
		w.Error = ""
	}
	if w.window != nil {
		w.window.Invalidate()
	}
}

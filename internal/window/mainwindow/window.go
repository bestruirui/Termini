package mainwindow

import (
	"log"

	"termini/internal/config"
	"termini/internal/platform/pty"
	"termini/internal/window/settingwindow"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// Open 创建主窗口并连接终端会话，窗口关闭时释放会话。
// preferences 和 settingWindow 必须非空，并由各窗口共享。
func Open(preferences *config.Settings, workingDirectory string, settingWindow *settingwindow.Window) error {
	var win *mygo.Window         // 承载终端的应用窗口。
	ready := make(chan struct{}) // 窗口创建后才允许终端回调访问窗口。
	command, err := pty.Command(preferences.Shell.Program, preferences.Shell.Arguments)
	if err != nil {
		return err
	}
	rightClick := terminal.RightClickCopyOrPaste // 当前终端会话使用的右键行为。
	if preferences.Shell.RightClickMenu {
		rightClick = terminal.RightClickContextMenu
	}
	cursor := terminal.CursorBlock // 当前终端会话的默认光标样式。
	switch preferences.Shell.CursorStyle {
	case "bar":
		cursor = terminal.CursorBar
	case "underline":
		cursor = terminal.CursorUnderline
	}
	term, err := pty.NewTerminal(terminal.Options{
		Command:         command,
		Dir:             workingDirectory,
		MinimumContrast: 3,
		RightClick:      rightClick,
		Cursor:          cursor,
		NoBlink:         !preferences.Shell.CursorBlink,
		OnTitle: func(title string) {
			<-ready
			if title == "" {
				title = "Termini"
			}
			win.SetTitle(title)
			win.Invalidate()
		},
		OnExit: func(int) { <-ready; win.Close() },
		OnNotify: func(title, body string) {
			if err := mygo.NewNotification(mygo.NotificationOptions{Title: title, Body: body}).Show(); err != nil {
				log.Print(err)
			}
		},
	}, preferences.Shell.Arguments)
	if err != nil {
		return err
	}
	term.SetFont(terminal.Font{Family: preferences.Shell.FontFamily, Size: float32(preferences.Shell.FontSize)})
	if err := settingwindow.ApplyAppearance(*preferences, term); err != nil {
		settingWindow.Error = err.Error()
		log.Print(err)
	}
	win = mygo.NewWindow(mygo.WindowOptions{
		Title:             "Termini",
		Width:             960,
		Height:            640,
		Hidden:            true,
		TitleBarStyle:     mygo.TitleBarHidden,
		HideWindowButtons: true,
		AutoHideMenuBar:   true,
		CornerRadius:      12,
		Content: ui.View(func(c *ui.Context) {
			theme := *c.Theme()
			theme.Font, theme.FontSize = preferences.Appearance.FontFamily, float32(preferences.Appearance.FontSize)
			c.SetTheme(&theme)
			ui.Column(c).Fill().Background(c.Theme().Surface).Children(func() {
				titleBar(c, win, "Termini", func() {
					settingWindow.Open(preferences, term, closeWindowButton)
				})
				ui.Row(c).Grow(1).MinHeight(0).AlignItems(ui.Stretch).Children(func() {
					pane := ui.Box(c).Grow(1).MinWidth(0).MinHeight(0)
					pane.Margin(0, 8, 8).Background(c.Theme().Background).Radius(12).Clip()
					pane.Children(func() { terminal.View(c, term).Fill().AutoFocus() })
				})
			})
		}),
	})
	win.OnReadyToShow(win.Show)
	close(ready)
	win.OnClosed(func() {
		term.Close()
	})
	return nil
}

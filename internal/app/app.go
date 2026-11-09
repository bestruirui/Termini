package app

import (
	"log"
	"os"
	"path/filepath"

	"termini/internal/config"
	"termini/internal/window/mainwindow"
	"termini/internal/window/settingwindow"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater/native"
)

// Run 启动桌面应用并运行事件循环，应从主 goroutine 调用。
func Run(workingDirectory string) error {
	dir, err := os.UserHomeDir() // 用户主目录，应用数据统一写入其中的 .termini。
	if err != nil {
		return err
	}
	dir = filepath.Join(dir, ".termini")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	mygo.App.SetPath(mygo.PathUserData, dir)
	preferences, err := config.Load() // 各窗口共享的用户设置，读取失败时使用默认值。
	settingWindow := settingwindow.Window{} // 独立设置窗口及其页面状态。
	if err != nil {
		settingWindow.Error = err.Error()
		log.Print(err)
	}
	mygo.Theme.SetSource(mygo.ThemeSource(preferences.Appearance.Mode))
	mygo.Use(native.Plugin)
	var startupErr error // 主线程创建主窗口时发生的启动错误。
	mygo.App.WhenReady(func() {
		startupErr = mainwindow.Open(&preferences, workingDirectory, &settingWindow)
		if startupErr != nil {
			mygo.App.Quit()
		}
	})
	if err := mygo.App.Run(); err != nil {
		return err
	}
	return startupErr
}

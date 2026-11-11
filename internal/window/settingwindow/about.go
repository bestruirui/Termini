package settingwindow

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/ui"
)

// aboutPage 显示版本和手动更新入口，编辑自动更新偏好并提供项目仓库链接。
func (w *Window) aboutPage(c *ui.Context) {
	version := mygo.App.Version() // 当前构建携带的应用版本，未打包时为空。
	group(c, "", func() {
		row(c, "当前版本 "+version, func() {
			if ui.Button(c.Key("check-updates"), "检查更新").Clicked() {
				updater.CheckForUpdates()
			}
		})
		ui.Divider(c)
		row(c, "自动更新", func() {
			automaticDownloads := updater.AutomaticDownloads() // 更新插件持久化的自动安装偏好。
			if ui.Switch(c.Key("auto-update"), &automaticDownloads).Label("自动更新").Changed() {
				updater.SetAutomaticDownloads(automaticDownloads)
			}
		})
	})
	group(c, "", func() {
		row(c, "GitHub", func() {
			if ui.Link(c.Key("repository"), "bestruirui/termini", "").SingleLine().Clicked() {
				if err := mygo.Shell.OpenExternal("https://github.com/bestruirui/termini"); err != nil {
					w.Error = "打开项目仓库失败：" + err.Error()
				} else {
					w.Error = ""
				}
			}
		})
	})
}

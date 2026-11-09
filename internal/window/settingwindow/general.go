package settingwindow

import (
	"runtime"

	"termini/internal/platform/explorer"

	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/ui"
)

// generalPage 编辑通用设置，开关变更后立即保存。
func (w *Window) generalPage(c *ui.Context) {
	group(c, "更新", func() {
		row(c, "自动更新", func() {
			automaticDownloads := updater.AutomaticDownloads() // 更新插件持久化的自动安装偏好。
			if ui.Switch(c.Key("auto-update"), &automaticDownloads).Label("自动更新").Changed() {
				updater.SetAutomaticDownloads(automaticDownloads)
			}
		})
		ui.Divider(c)
		row(c, "检查更新", func() {
			automaticChecks := updater.AutomaticChecks() // 更新插件持久化的后台检查偏好。
			if ui.Switch(c.Key("check-updates"), &automaticChecks).Label("检查更新").Changed() {
				updater.SetAutomaticChecks(automaticChecks)
			}
		})
	})
	if runtime.GOOS == "windows" {
		group(c, "右键菜单", func() {
			row(c, "使用 Termini 打开", func() {
				if ui.Switch(c.Key("explorer-context-menu"), &w.preferences.General.ExplorerContextMenu).Label("使用 Termini 打开").Changed() {
					if err := explorer.SetContextMenu(w.preferences.General.ExplorerContextMenu); err != nil {
						w.preferences.General.ExplorerContextMenu = !w.preferences.General.ExplorerContextMenu
						w.Error = "设置右键菜单失败：" + err.Error()
					} else {
						w.save()
					}
				}
			})
		})
	}
}

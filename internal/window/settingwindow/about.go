package settingwindow

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// aboutPage 显示项目仓库，点击后使用系统默认浏览器打开。
func (w *Window) aboutPage(c *ui.Context) {
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

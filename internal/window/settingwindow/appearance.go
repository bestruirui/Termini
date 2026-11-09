package settingwindow

import (
	"slices"

	"termini/internal/component"
	"termini/internal/config"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// appearancePage 编辑深浅模式、终端配色和应用文字，修改后立即生效。
func (w *Window) appearancePage(c *ui.Context) {
	changed := false // 本次界面构建是否收到外观设置变更。
	group(c, "配色", func() {
		row(c, "深浅模式", func() {
			modes := []string{"system", "light", "dark"}
			labels := []string{"跟随系统", "浅色", "深色"}
			mode := labels[slices.Index(modes, w.preferences.Appearance.Mode)]
			if component.Select(c.Key("appearance"), &mode, labels, 150).Changed() {
				w.preferences.Appearance.Mode = modes[slices.Index(labels, mode)]
				changed = true
			}
		})
		ui.Divider(c)
		row(c, "终端主题", func() {
			theme := w.preferences.Appearance.Theme
			if theme == "" {
				theme = "默认"
			}
			if component.SearchSelect(c.Key("theme"), &theme, w.themeNames, 240) {
				if theme == "默认" {
					theme = ""
				}
				w.preferences.Appearance.Theme = theme
				changed = true
			}
		})
	})
	group(c, "文字", func() {
		row(c, "应用字体", func() {
			font := w.preferences.Appearance.FontFamily
			if font == "" {
				font = "系统字体"
			}
			if component.SearchSelect(c.Key("font"), &font, w.appFontNames, 240) {
				if font == "系统字体" {
					font = ""
				}
				w.preferences.Appearance.FontFamily = font
				changed = true
			}
		})
		ui.Divider(c)
		row(c, "字号", func() {
			var sizes []float64
			for size := 8; size <= 48; size++ {
				sizes = append(sizes, float64(size))
			}
			if !slices.Contains(sizes, w.preferences.Appearance.FontSize) {
				sizes = append(sizes, w.preferences.Appearance.FontSize)
				slices.Sort(sizes)
			}
			changed = component.Select(c.Key("font-size"), &w.preferences.Appearance.FontSize, sizes, 90).Changed() || changed
		})
	})
	if changed {
		if err := ApplyAppearance(*w.preferences, w.term); err != nil {
			w.Error = err.Error()
		} else {
			w.save()
		}
	}
}

// ApplyAppearance 应用深浅模式和终端配色，并刷新所有窗口以显示当前应用文字设置。
func ApplyAppearance(preferences config.Settings, term *terminal.Terminal) error {
	var theme *terminal.Theme // 空值让终端按窗口的深浅模式选择默认配色。
	if preferences.Appearance.Theme != "" {
		var err error
		theme, err = terminal.GhosttyTheme(preferences.Appearance.Theme)
		if err != nil {
			return err
		}
	}
	mygo.Theme.SetSource(mygo.ThemeSource(preferences.Appearance.Mode))
	if term != nil {
		term.SetTheme(theme, nil)
	}
	for _, window := range mygo.Windows() {
		window.Invalidate()
	}
	return nil
}

package component

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// searchMenu 保存可搜索下拉菜单的展开、筛选和键盘导航状态。
type searchMenu struct {
	open    bool    // 菜单当前是否展开。
	query   string  // 搜索框中的筛选文字。
	hi      int     // 当前高亮项在筛选结果中的位置。
	hovered string  // 鼠标最近进入的选项，用于区分鼠标与键盘导航。
	reveal  bool    // 本帧是否需要将高亮项滚动到可见区域。
	at      ui.Rect // 菜单展开时按钮的位置，按钮移动后关闭菜单。
}

var (
	searchIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 56 56"><path fill="currentColor" d="M23.957 41.77a18.02 18.02 0 0 0 10.477-3.376l11.109 11.11a2.66 2.66 0 0 0 1.898.773c1.524 0 2.625-1.172 2.625-2.672c0-.703-.234-1.359-.75-1.874L38.277 34.668c2.32-3.047 3.703-6.82 3.703-10.922c0-9.914-8.109-18.023-18.023-18.023c-9.937 0-18.023 8.109-18.023 18.023S14.02 41.77 23.957 41.77m0-3.891c-7.758 0-14.133-6.398-14.133-14.133S16.2 9.613 23.957 9.613c7.734 0 14.133 6.399 14.133 14.133c0 7.735-6.399 14.133-14.133 14.133"/></svg>`)) // 下拉菜单搜索框的放大镜图标。
	upDownIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="m7 15l5 5l5-5M7 9l5-5l5 5"/></svg>`)) // 下拉控件的上下箭头。
	checkIcon  = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M20 6L9 17l-5-5"/></svg>`)) // 下拉菜单当前选项的勾选标记。
)

// Select 绘制紧凑的无边框下拉控件，展开后让当前选项与按钮文字对齐。
func Select[T comparable](c *ui.Context, value *T, options []T, width float32) ui.Element {
	sel := ui.SelectBase(c, value)
	trigger := sel.Trigger.MaxWidth(width).Label(fmt.Sprint(*value))
	popupFace(c, trigger, fmt.Sprint(*value), sel.Open())
	var menuWidth float32 // 本帧确定的菜单宽度，不依赖上一帧的布局结果。
	var spot ui.Element   // 菜单使用窗口坐标对齐的零尺寸锚点。
	var maxHeight float32 // 窗口上下各留出 8 像素后的菜单可用高度。
	if sel.Open() {
		bounds := trigger.Bounds()
		windowWidth, windowHeight := c.Size()
		menuWidth = bounds.W
		// 首帧按最长选项确定宽度，45 像素包含边距、勾选图标及其间距。
		for _, option := range options {
			textWidth, _ := c.MeasureText(0, ui.Span{Text: fmt.Sprint(option), Size: c.Theme().FontSize})
			menuWidth = max(menuWidth, textWidth+45)
		}
		menuWidth = min(menuWidth, windowWidth-16)
		maxHeight = windowHeight - 16
		itemHeight := max(float32(24), c.Theme().FontSize+11)
		menuHeight := min(10+float32(len(options))*itemHeight, maxHeight)
		// 抵消菜单内边距和勾选区后对齐文字，并让当前项中心覆盖按钮中心。
		x := max(float32(8), min(bounds.X-17, windowWidth-8-menuWidth))
		y := max(float32(8), min(bounds.Y-(3+float32(max(slices.Index(options, *value), 0))*itemHeight), windowHeight-8-menuHeight))
		ui.Overlay(c.Key("menu-spot"), func() {
			spot = ui.Box(c).Absolute().Left(x).Top(y).Size(0, 0).PassThrough()
		})
	}
	sel.Popup(func(panel ui.Element) {
		styleMenu(c, panel)
		panel.Width(menuWidth).AttachTo(spot, ui.AnchorTopLeft, ui.AnchorTopLeft)
		ui.Scroll(c.Key("items")).MaxHeight(maxHeight - 10).Children(func() {
			for _, option := range options {
				item := sel.Item(option)
				// 鼠标离开菜单时保留键盘导航高亮，当前值只显示勾选标记。
				highlighted := item.Highlighted() && (item.Hovered() || !panel.Hovered() && option != *value)
				menuRow(c, item, fmt.Sprint(option), option == *value, highlighted)
			}
		})
	})
	return trigger
}

// SearchSelect 在按钮附近展开带搜索框的菜单，返回本帧是否选择了新值。
func SearchSelect(c *ui.Context, value *string, options []string, width float32) bool {
	changed := false // 本帧的菜单操作是否修改了选中值。
	trigger := ui.ButtonBase(c).MaxWidth(width).Label(*value)
	state := ui.Local(trigger, "menu", func() searchMenu { return searchMenu{} }) // 此控件跨帧保留的菜单状态。
	if trigger.Clicked() {
		state.open = !state.open
		state.query, state.hovered, state.reveal = "", "", true
		state.hi = max(slices.Index(options, *value), 0)
		state.at = trigger.Bounds()
	}
	// 按钮因窗口尺寸或布局变化而移动时，关闭原位置的菜单。
	if bounds := trigger.Bounds(); state.open && (math.Abs(float64(bounds.X-state.at.X)) > 0.5 || math.Abs(float64(bounds.Y-state.at.Y)) > 0.5) {
		state.open = false
	}
	if state.open {
		ui.Overlay(c.Key("menu-backdrop"), func() {
			back := ui.ButtonBase(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Label("关闭菜单")
			if back.Clicked() {
				state.open = false
			}
		})
	}
	popupFace(c, trigger, *value, state.open)
	bounds := trigger.Bounds()
	_, windowHeight := c.Size()
	room := max(windowHeight-(bounds.Y+bounds.H)-12, bounds.Y-12)
	ui.PopoverBase(c.Key("menu"), trigger, &state.open, func(panel ui.Element) {
		styleMenu(c, panel)
		panel.Margin(4, 0, 0, 0).Width(max(width, 260))
		var input ui.Element // 搜索框的输入节点，负责筛选和键盘导航。
		inputHeight := max(float32(26), c.Theme().FontSize+13)
		frame := ui.Row(c.Key("search")).Height(inputHeight).Padding(0, 8).Gap(6).AlignItems(ui.Center).Margin(3, 3, 8, 3).Cursor(ui.CursorText)
		frame.Children(func() {
			ui.Icon(c, searchIcon).Size(12, 12).TextColor(c.Theme().TextMuted).Shrink(0)
			input = ui.TextInputBase(c.Key("query"), &state.query).AutoFocus().Grow(1).MinWidth(0).AlignSelf(ui.Stretch)
			input.Placeholder("搜索").Label("搜索").FocusRing(false)
		})
		styleInput(c, frame, input.Focused())
		query := strings.ToLower(strings.TrimSpace(state.query))
		var shown []string
		for _, option := range options {
			if query == "" || strings.Contains(strings.ToLower(option), query) {
				shown = append(shown, option)
			}
		}
		state.hi = min(max(state.hi, 0), max(len(shown)-1, 0))
		if input.Changed() {
			state.hi, state.reveal = 0, true
		}
		input.OnShortcut(0, ui.KeyDown, func() {
			state.hi, state.reveal = min(state.hi+1, max(len(shown)-1, 0)), true
		})
		input.OnShortcut(0, ui.KeyUp, func() {
			state.hi, state.reveal = max(state.hi-1, 0), true
		})
		pick := func(index int) {
			if index >= 0 && index < len(shown) {
				state.open = false
				if shown[index] != *value {
					*value, changed = shown[index], true
				}
			}
		}
		if input.Submitted() {
			pick(state.hi)
		}
		if len(shown) == 0 {
			ui.Text(c, "无匹配结果").TextColor(c.Theme().TextMuted).Padding(6, 12)
			return
		}
		reveal := state.reveal
		state.reveal = false
		moved := -1 // 本帧鼠标新进入的选项，统一在列表构建后更新高亮。
		// 扣除搜索框高度与 21 像素边距后，列表至少容纳一项且最多使用 300 像素。
		ui.Scroll(c.Key("items")).MaxHeight(max(max(float32(24), c.Theme().FontSize+11), min(300, room-inputHeight-21))).Children(func() {
			for i, option := range shown {
				item := ui.Row(c.Key(option)).Label(option)
				if item.Hovered() && state.hovered != option {
					moved = i
				}
				if item.Clicked() {
					pick(i)
				}
				menuRow(c, item, option, option == *value, i == state.hi)
				if reveal && i == state.hi {
					item.ScrollIntoView()
				}
			}
		})
		if moved >= 0 {
			state.hovered, state.hi = shown[moved], moved
			c.After(0)
		}
	})
	return changed
}

// popupFace 绘制下拉按钮的文字、圆形箭头底色及展开状态。
func popupFace(c *ui.Context, trigger ui.Element, text string, open bool) {
	// 收起按钮与 Switch 等高，圆形箭头适配扣除上下内边距后的高度。
	trigger.Height(c.Theme().Space(5)).Padding(c.Theme().Space(0.5), 3, c.Theme().Space(0.5), 10).Gap(6).Radius(c.Theme().Radius).AlignItems(ui.Center).Cursor(ui.CursorDefault)
	background, disc := ui.Transparent, c.Theme().SurfaceHover // 按下或展开时由整个控件承接箭头底色。
	switch {
	case open || trigger.Pressed():
		background, disc = c.Theme().SurfacePressed, ui.Transparent
	case trigger.Hovered():
		background, disc = c.Theme().SurfaceHover, ui.Transparent
		trigger.Shadow(0, 0, 0, c.Theme().Space(0.5), c.Theme().SurfaceHover)
	}
	trigger.Background(trigger.AnimateColor("bg", background, ui.Animation{Duration: 120 * time.Millisecond}))
	trigger.Children(func() {
		ui.Text(c, text).TextColor(c.Theme().Text).SingleLine().Ellipsis("…").MinWidth(0).Shrink(1)
		arrow := ui.Box(c.Key("disc")).Size(c.Theme().Space(4), c.Theme().Space(4)).Radius(c.Theme().Space(2)).Center().Shrink(0)
		arrow.Background(arrow.AnimateColor("bg", disc, ui.Animation{Duration: 120 * time.Millisecond})).Children(func() {
			ui.Icon(c, upDownIcon).Size(c.Theme().Space(3), c.Theme().Space(3)).TextColor(c.Theme().Text.Alpha(0.85))
		})
	})
}

// styleMenu 为普通菜单与搜索菜单设置相同的圆角、边线和浮层底色。
func styleMenu(c *ui.Context, panel ui.Element) {
	panel.Padding(5).Radius(13).Background(c.Theme().Background).InsetShadow(0, 0, 0, 1, c.Theme().Border)
	panel.Shadow(0, 5, 10, 0, ui.RGBA(0, 0, 0, 0.22))
}

// menuRow 绘制菜单项的勾选标记、文字和导航高亮。
func menuRow(c *ui.Context, item ui.Element, label string, current, highlighted bool) {
	item.Height(max(float32(24), c.Theme().FontSize+11)).Padding(0, 12, 0, 6).Gap(5).Radius(7).AlignItems(ui.Center).TextColor(c.Theme().Text)
	if highlighted {
		item.Background(c.Theme().Accent).TextColor(c.Theme().AccentText)
	}
	item.Children(func() {
		mark := ui.Icon(c, checkIcon).Size(12, 12).Shrink(0)
		if !current {
			mark.Opacity(0)
		}
		ui.Text(c, label).SingleLine().Ellipsis("…").MinWidth(0)
	})
}

// styleInput 为文本输入添加主题底色、内边线和渐变显示的聚焦轮廓。
func styleInput(c *ui.Context, input ui.Element, focused bool) {
	input.Radius(7).FocusRing(false)
	input.Background(c.Theme().Surface).InsetShadow(0, 0, 0, 1, c.Theme().Border)
	focus := float32(0) // 聚焦轮廓的目标显现程度。
	if focused {
		focus = 1
	}
	glow := input.Animate("glow", focus, 120*time.Millisecond)
	if glow > 0 {
		focusColor := c.Theme().Focus.Alpha(float32(c.Theme().Focus.A) / 255 * glow) // 保留主题轮廓色的不透明度并随聚焦动画淡入。
		input.Shadow(0, 0, 0, 3.5*glow, focusColor.Alpha(float32(focusColor.A)/255*0.5)).InsetShadow(0, 0, 0, 1, focusColor)
	}
}

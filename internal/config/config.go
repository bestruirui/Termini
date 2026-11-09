package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Settings 按通用、外观和终端页面组织用户偏好。
type Settings struct {
	General struct {
		ExplorerContextMenu bool `json:"explorer_context_menu"` // 文件管理器右键菜单开关，默认关闭。
	} `json:"general"` // 通用页的右键菜单设置。
	Appearance struct {
		Mode       string  `json:"mode"`        // 应用深浅模式：system、light 或 dark。
		FontFamily string  `json:"font_family"` // 应用字体族，空值使用系统字体。
		FontSize   float64 `json:"font_size"`   // 应用基准字号，单位为 DIP。
		Theme      string  `json:"theme"`       // 终端配色主题，空值使用默认配色。
	} `json:"appearance"` // 外观页的配色与应用文字设置。
	Shell struct {
		FontFamily     string  `json:"font_family"`      // 终端字体族，空值使用系统等宽字体。
		FontSize       float64 `json:"font_size"`        // 终端字号，单位为 DIP。
		CursorStyle    string  `json:"cursor_style"`     // 默认光标样式：bar、block 或 underline，终端程序可自行改变。
		CursorBlink    bool    `json:"cursor_blink"`     // 默认光标是否闪烁，默认开启，终端程序可自行改变。
		RightClickMenu bool    `json:"right_click_menu"` // 右键是否显示菜单，默认有选区时复制、无选区时粘贴。
		Program        string  `json:"program"`          // 启动程序的路径或名称，空值自动选择 shell。
		Arguments      string  `json:"arguments"`        // 原样保存并传入 shell 的单行启动参数。
	} `json:"shell"` // 终端页的文字、光标、鼠标和启动设置。
}

// Load 从 ~/.termini 目录读取设置，文件不存在时返回默认值。
func Load() (Settings, error) {
	defaults := Settings{} // 首次启动及缺失字段使用的默认设置。
	defaults.Appearance.Mode = "system"
	defaults.Appearance.FontSize, defaults.Shell.FontSize = 13, 13
	defaults.Shell.CursorStyle, defaults.Shell.CursorBlink = "block", true
	dir, err := os.UserHomeDir()
	if err != nil {
		return defaults, err
	}
	data, err := os.ReadFile(filepath.Join(dir, ".termini", "settings.json"))
	if errors.Is(err, os.ErrNotExist) {
		return defaults, nil
	}
	if err != nil {
		return defaults, err
	}
	s := defaults
	if err := json.Unmarshal(data, &s); err != nil {
		return defaults, fmt.Errorf("读取设置失败：%w", err)
	}
	if s.Appearance.Mode != "system" && s.Appearance.Mode != "light" && s.Appearance.Mode != "dark" {
		return defaults, errors.New("设置中的深浅模式无效")
	}
	if s.Appearance.FontSize < 8 || s.Appearance.FontSize > 48 {
		return defaults, errors.New("设置中的应用字号须在 8 至 48 之间")
	}
	if s.Shell.FontSize < 8 || s.Shell.FontSize > 48 {
		return defaults, errors.New("设置中的字号须在 8 至 48 之间")
	}
	if s.Shell.CursorStyle != "bar" && s.Shell.CursorStyle != "block" && s.Shell.CursorStyle != "underline" {
		return defaults, errors.New("设置中的光标样式无效")
	}
	return s, nil
}

// Save 将设置写入 ~/.termini 目录，供下次启动恢复。
func Save(s Settings) error {
	dir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir = filepath.Join(dir, ".termini")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "settings.json"), data, 0o600)
}

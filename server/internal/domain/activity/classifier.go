// Package activity 活动领域。
//
// classifier.go — 活动分类领域服务
//
// 纯同步业务规则：根据应用名、窗口标题、浏览器 URL 判定活动类别。
// 先通过应用名映射表确定基础分类，再结合标题/URL 推断语义分类。
// 从 Rust classifier_service.rs 1:1 移植。
package activity

import "strings"

// Classify 对一条活动进行分类
func Classify(appName, windowTitle string, browserURL *string) ClassificationResult {
	category := classifyByApp(appName)
	semantic, confidence := classifySemantic(category, windowTitle, browserURL)
	return ClassificationResult{
		Category:         category,
		SemanticCategory: semantic,
		Confidence:       confidence,
	}
}

// classifyByApp 根据应用名确定基础分类
func classifyByApp(appName string) string {
	name := strings.ToLower(appName)

	// 编码开发
	codingApps := map[string]bool{
		"code": true, "cursor": true, "idea": true, "intellij": true,
		"pycharm": true, "webstorm": true, "goland": true, "clion": true,
		"rider": true, "rustrover": true, "datagrip": true,
		"sublime_text": true, "sublime text": true,
		"notepad++": true, "vim": true, "neovim": true, "nvim": true,
		"emacs": true, "helix": true, "zed": true, "fleet": true,
		"android studio": true, "xcode": true,
	}
	if codingApps[name] {
		return CategoryCoding
	}

	// 浏览器
	browserApps := map[string]bool{
		"chrome": true, "msedge": true, "edge": true, "firefox": true,
		"brave": true, "opera": true, "vivaldi": true, "arc": true,
		"safari": true, "chromium": true,
	}
	if browserApps[name] {
		return CategoryBrowser
	}

	// 即时通讯
	commApps := map[string]bool{
		"wechat": true, "微信": true, "qq": true, "dingtalk": true, "钉钉": true,
		"feishu": true, "飞书": true, "lark": true, "slack": true, "teams": true,
		"discord": true, "telegram": true, "zoom": true, "tencent meeting": true,
	}
	if commApps[name] {
		return CategoryCommunication
	}

	// 文档编辑
	docApps := map[string]bool{
		"word": true, "winword": true, "excel": true, "powerpoint": true,
		"powerpnt": true, "onenote": true, "notion": true, "obsidian": true,
		"typora": true, "wps": true, "wps office": true,
		"pages": true, "numbers": true, "keynote": true,
	}
	if docApps[name] {
		return CategoryDocument
	}

	// 终端
	termApps := map[string]bool{
		"windowsterminal": true, "windows terminal": true,
		"cmd": true, "powershell": true, "pwsh": true,
		"warp": true, "alacritty": true, "wezterm": true, "kitty": true,
		"iterm2": true, "terminal": true, "hyper": true,
	}
	if termApps[name] {
		return CategoryTerminal
	}

	// 设计
	designApps := map[string]bool{
		"figma": true, "sketch": true, "photoshop": true, "illustrator": true,
		"canva": true, "affinity": true, "inkscape": true, "gimp": true,
		"blender": true, "after effects": true,
	}
	if designApps[name] {
		return CategoryDesign
	}

	// 媒体
	mediaApps := map[string]bool{
		"spotify": true, "vlc": true, "网易云音乐": true, "cloudmusic": true,
		"qqmusic": true, "potplayer": true, "mpv": true,
		"foobar2000": true, "musicbee": true,
	}
	if mediaApps[name] {
		return CategoryMedia
	}

	// 系统工具
	sysApps := map[string]bool{
		"explorer": true, "finder": true, "task manager": true,
		"taskmgr": true, "devenv": true, "regedit": true,
		"everything": true, "7zip": true, "winrar": true,
	}
	if sysApps[name] {
		return CategorySystem
	}

	// 游戏
	gameApps := map[string]bool{
		"steam": true, "epic games": true, "origin": true,
		"battle.net": true, "genshin impact": true,
	}
	if gameApps[name] {
		return CategoryGaming
	}

	return CategoryOther
}

// classifySemantic 结合标题/URL 推断语义分类
func classifySemantic(baseCategory, windowTitle string, browserURL *string) (string, int) {
	titleLower := strings.ToLower(windowTitle)

	// 浏览器需要进一步细分
	if baseCategory == CategoryBrowser {
		return classifyBrowserSemantic(titleLower, browserURL)
	}

	// 其他应用直接从基础分类映射
	switch baseCategory {
	case CategoryCoding, CategoryTerminal:
		return SemanticCoding, 85
	case CategoryCommunication:
		return SemanticCommunication, 85
	case CategoryDocument:
		return SemanticWriting, 85
	case CategoryDesign:
		return SemanticDesign, 85
	case CategoryMedia, CategoryGaming:
		return SemanticEntertainment, 85
	case CategorySystem:
		return SemanticSystemOperation, 85
	default:
		return SemanticOther, 85
	}
}

// classifyBrowserSemantic 浏览器语义分类
func classifyBrowserSemantic(titleLower string, browserURL *string) (string, int) {
	urlLower := ""
	if browserURL != nil {
		urlLower = strings.ToLower(*browserURL)
	}

	// 编码相关
	if hasAnyKeyword(urlLower, []string{
		"github.com", "gitlab.com", "stackoverflow.com",
		"stackexchange.com", "crates.io", "docs.rs",
		"npmjs.com", "pypi.org", "pkg.go.dev",
	}) {
		return SemanticCoding, 90
	}

	// 调研/学习
	if hasAnyKeyword(urlLower, []string{
		"wikipedia.org", "zhihu.com", "juejin.cn",
		"csdn.net", "cnblogs.com", "segmentfault.com",
		"medium.com", "dev.to", "arxiv.org",
	}) || hasAnyKeyword(titleLower, []string{
		"教程", "文档", "tutorial", "guide", "reference",
	}) {
		return SemanticResearch, 80
	}

	// 沟通
	if hasAnyKeyword(urlLower, []string{
		"mail.google.com", "outlook.live.com",
		"mail.qq.com", "web.telegram.org",
		"web.whatsapp.com",
	}) {
		return SemanticCommunication, 85
	}

	// 娱乐
	if hasAnyKeyword(urlLower, []string{
		"youtube.com", "bilibili.com", "netflix.com",
		"twitch.tv", "v.qq.com", "iqiyi.com",
		"douyin.com", "weibo.com",
	}) {
		return SemanticEntertainment, 80
	}

	// 默认浏览
	return SemanticBrowsing, 60
}

// hasAnyKeyword 检查文本是否包含任一关键词
func hasAnyKeyword(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

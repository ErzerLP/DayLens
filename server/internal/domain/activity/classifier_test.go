package activity

import "testing"

// IDE 应用应分类为 coding
func TestClassify_IDEShouldReturnCoding(t *testing.T) {
	result := Classify("Code", "main.rs", nil)
	if result.Category != CategoryCoding {
		t.Errorf("got %s, want %s", result.Category, CategoryCoding)
	}
	if result.SemanticCategory != SemanticCoding {
		t.Errorf("got %s, want %s", result.SemanticCategory, SemanticCoding)
	}
}

// 浏览器应用应分类为 browser
func TestClassify_BrowserShouldReturnBrowser(t *testing.T) {
	url := "https://www.google.com"
	result := Classify("Chrome", "Google", &url)
	if result.Category != CategoryBrowser {
		t.Errorf("got %s, want %s", result.Category, CategoryBrowser)
	}
}

// GitHub 页面应语义分类为 Coding
func TestClassify_GitHubShouldSemanticCoding(t *testing.T) {
	url := "https://github.com/rust-lang/rust"
	result := Classify("Chrome", "rust-lang/rust - GitHub", &url)
	if result.Category != CategoryBrowser {
		t.Errorf("category: got %s, want %s", result.Category, CategoryBrowser)
	}
	if result.SemanticCategory != SemanticCoding {
		t.Errorf("semantic: got %s, want %s", result.SemanticCategory, SemanticCoding)
	}
	if result.Confidence < 80 {
		t.Errorf("confidence %d < 80", result.Confidence)
	}
}

// Bilibili 应语义分类为 Entertainment
func TestClassify_BilibiliShouldSemanticEntertainment(t *testing.T) {
	url := "https://www.bilibili.com/video/123"
	result := Classify("Chrome", "哔哩哔哩", &url)
	if result.SemanticCategory != SemanticEntertainment {
		t.Errorf("got %s, want %s", result.SemanticCategory, SemanticEntertainment)
	}
}

// 即时通讯应用应分类为 communication
func TestClassify_WeChatShouldReturnCommunication(t *testing.T) {
	result := Classify("WeChat", "聊天", nil)
	if result.Category != CategoryCommunication {
		t.Errorf("category: got %s, want %s", result.Category, CategoryCommunication)
	}
	if result.SemanticCategory != SemanticCommunication {
		t.Errorf("semantic: got %s, want %s", result.SemanticCategory, SemanticCommunication)
	}
}

// 终端应分类为 terminal，语义为 Coding
func TestClassify_TerminalShouldReturnTerminalAndCoding(t *testing.T) {
	result := Classify("WindowsTerminal", "pwsh", nil)
	if result.Category != CategoryTerminal {
		t.Errorf("category: got %s, want %s", result.Category, CategoryTerminal)
	}
	if result.SemanticCategory != SemanticCoding {
		t.Errorf("semantic: got %s, want %s", result.SemanticCategory, SemanticCoding)
	}
}

// 未知应用应分类为 other
func TestClassify_UnknownAppShouldReturnOther(t *testing.T) {
	result := Classify("SomeRandomApp", "", nil)
	if result.Category != CategoryOther {
		t.Errorf("got %s, want %s", result.Category, CategoryOther)
	}
}

// 分类应不区分大小写
func TestClassify_ShouldBeCaseInsensitive(t *testing.T) {
	result := Classify("CHROME", "Test", nil)
	if result.Category != CategoryBrowser {
		t.Errorf("got %s, want %s", result.Category, CategoryBrowser)
	}
}

// 知乎应语义分类为 Research
func TestClassify_ZhihuShouldSemanticResearch(t *testing.T) {
	url := "https://www.zhihu.com/question/12345"
	result := Classify("Chrome", "如何学习 Go 语言", &url)
	if result.SemanticCategory != SemanticResearch {
		t.Errorf("got %s, want %s", result.SemanticCategory, SemanticResearch)
	}
}

// 文档应用应分类为 document，语义为 Writing
func TestClassify_DocumentAppShouldReturnWriting(t *testing.T) {
	result := Classify("Notion", "项目方案", nil)
	if result.Category != CategoryDocument {
		t.Errorf("category: got %s, want %s", result.Category, CategoryDocument)
	}
	if result.SemanticCategory != SemanticWriting {
		t.Errorf("semantic: got %s, want %s", result.SemanticCategory, SemanticWriting)
	}
}

// 设计应用应分类为 design
func TestClassify_DesignAppShouldReturnDesign(t *testing.T) {
	result := Classify("Figma", "UI Design", nil)
	if result.Category != CategoryDesign {
		t.Errorf("got %s, want %s", result.Category, CategoryDesign)
	}
	if result.SemanticCategory != SemanticDesign {
		t.Errorf("semantic: got %s, want %s", result.SemanticCategory, SemanticDesign)
	}
}

// 媒体应用应分类为 media，语义为 Entertainment
func TestClassify_MediaAppShouldReturnEntertainment(t *testing.T) {
	result := Classify("Spotify", "Daily Mix", nil)
	if result.Category != CategoryMedia {
		t.Errorf("got %s, want %s", result.Category, CategoryMedia)
	}
	if result.SemanticCategory != SemanticEntertainment {
		t.Errorf("semantic: got %s, want %s", result.SemanticCategory, SemanticEntertainment)
	}
}

// Gmail 应语义分类为 Communication
func TestClassify_GmailShouldSemanticCommunication(t *testing.T) {
	url := "https://mail.google.com/mail/u/0/"
	result := Classify("Chrome", "Inbox - Gmail", &url)
	if result.SemanticCategory != SemanticCommunication {
		t.Errorf("got %s, want %s", result.SemanticCategory, SemanticCommunication)
	}
}

// 普通网站应语义分类为 Browsing
func TestClassify_GenericSiteShouldSemanticBrowsing(t *testing.T) {
	url := "https://www.example.com"
	result := Classify("Chrome", "Example", &url)
	if result.SemanticCategory != SemanticBrowsing {
		t.Errorf("got %s, want %s", result.SemanticCategory, SemanticBrowsing)
	}
}

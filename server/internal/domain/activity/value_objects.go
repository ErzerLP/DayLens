package activity

// Category 基础分类常量 — 10 种
const (
	CategoryCoding        = "coding"
	CategoryBrowser       = "browser"
	CategoryCommunication = "communication"
	CategoryDocument      = "document"
	CategoryDesign        = "design"
	CategoryTerminal      = "terminal"
	CategoryMedia         = "media"
	CategorySystem        = "system"
	CategoryGaming        = "gaming"
	CategoryOther         = "other"
)

// AllCategories 所有基础分类
var AllCategories = []string{
	CategoryCoding, CategoryBrowser, CategoryCommunication,
	CategoryDocument, CategoryDesign, CategoryTerminal,
	CategoryMedia, CategorySystem, CategoryGaming, CategoryOther,
}

// SemanticCategory 语义分类常量 — 9 种
const (
	SemanticCoding          = "Coding"
	SemanticResearch        = "Research"
	SemanticCommunication   = "Communication"
	SemanticWriting         = "Writing"
	SemanticDesign          = "Design"
	SemanticBrowsing        = "Browsing"
	SemanticEntertainment   = "Entertainment"
	SemanticSystemOperation = "SystemOperation"
	SemanticOther           = "Other"
)

// AllSemanticCategories 所有语义分类
var AllSemanticCategories = []string{
	SemanticCoding, SemanticResearch, SemanticCommunication,
	SemanticWriting, SemanticDesign, SemanticBrowsing,
	SemanticEntertainment, SemanticSystemOperation, SemanticOther,
}

// ClassificationResult 分类结果 — 由 ClassifierService 返回
type ClassificationResult struct {
	Category           string
	SemanticCategory   string
	Confidence         int
}

// CategoryRule 自定义分类规则
type CategoryRule struct {
	AppName  string `json:"appName"`
	Category string `json:"category"`
}

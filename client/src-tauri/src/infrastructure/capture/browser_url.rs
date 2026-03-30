//! # 浏览器 URL 提取
//!
//! 通过 Windows UI Automation API 提取浏览器地址栏中的 URL。
//! 支持 Chrome、Edge、Brave（Chromium 系）和 Firefox。
//!
//! # 工作原理
//! 1. `CoCreateInstance` 创建 `IUIAutomation` COM 对象
//! 2. `ElementFromHandle(HWND)` 获取浏览器窗口的自动化元素
//! 3. 根据浏览器类型查找地址栏控件（Edit / 自定义控件）
//! 4. 读取控件的 `Value` 属性获取 URL
//!
//! # 线程安全
//! COM 对象不实现 Send/Sync，因此每次调用时创建新实例。
//! 在采集循环中调用频率不高（30秒一次），开销可接受。

use windows::core::Interface;
use windows::Win32::Foundation::HWND;
use windows::Win32::System::Com::{
    CoCreateInstance, CoInitializeEx, CLSCTX_ALL,
    COINIT_MULTITHREADED,
};
use windows::Win32::UI::Accessibility::{
    CUIAutomation, IUIAutomation, IUIAutomationCondition,
    IUIAutomationElement, TreeScope_Subtree, UIA_EditControlTypeId,
    UIA_ValuePatternId,
};

// ===== BrowserUrlExtractor =====

/// 浏览器 URL 提取器
///
/// 通过 UI Automation 读取浏览器地址栏的 URL。
/// 每次调用创建新的 COM 实例（COM 对象非 Send/Sync）。
pub struct BrowserUrlExtractor;

impl BrowserUrlExtractor {
    /// 创建 IUIAutomation 实例
    ///
    /// 每次调用都会初始化 COM 并创建新实例。
    fn create_uia() -> Option<IUIAutomation> {
        unsafe {
            let _ = CoInitializeEx(None, COINIT_MULTITHREADED);
            CoCreateInstance::<_, IUIAutomation>(
                &CUIAutomation,
                None,
                CLSCTX_ALL,
            )
            .ok()
        }
    }

    /// 从前台浏览器窗口提取 URL
    ///
    /// # 参数
    /// - `hwnd` — 浏览器窗口句柄
    /// - `app_name` — 应用名（如 "Chrome", "msedge", "firefox"）
    ///
    /// # 返回
    /// - `Some(url)` — 成功提取 URL
    /// - `None` — 提取失败或非浏览器窗口
    pub fn extract(hwnd: HWND, app_name: &str) -> Option<String> {
        let uia = Self::create_uia()?;
        let name_lower = app_name.to_lowercase();

        if is_chromium_browser(&name_lower) {
            Self::extract_chromium_url(&uia, hwnd)
        } else if name_lower == "firefox" {
            Self::extract_firefox_url(&uia, hwnd)
        } else {
            None
        }
    }

    /// 提取 Chromium 系浏览器 URL（Chrome / Edge / Brave / Vivaldi）
    ///
    /// Chromium 地址栏是一个 Edit 控件，
    /// 通过 `UIA_EditControlTypeId` 条件查找。
    fn extract_chromium_url(
        uia: &IUIAutomation,
        hwnd: HWND,
    ) -> Option<String> {
        unsafe {
            let root = uia.ElementFromHandle(hwnd).ok()?;
            let condition = create_edit_condition(uia)?;
            let element = find_url_edit_element(&root, &condition)?;
            read_value_pattern(&element)
        }
    }

    /// 提取 Firefox URL
    ///
    /// Firefox 的 UI 结构与 Chromium 不同，
    /// 但地址栏同样可通过 Edit 控件查找。
    fn extract_firefox_url(
        uia: &IUIAutomation,
        hwnd: HWND,
    ) -> Option<String> {
        unsafe {
            let root = uia.ElementFromHandle(hwnd).ok()?;
            let condition = create_edit_condition(uia)?;
            let element = find_url_edit_element(&root, &condition)?;
            read_value_pattern(&element)
        }
    }
}

// ===== 辅助函数 =====

/// 判断是否为 Chromium 系浏览器
fn is_chromium_browser(name_lower: &str) -> bool {
    matches!(
        name_lower,
        "chrome" | "msedge" | "brave" | "opera"
        | "vivaldi" | "arc" | "chromium"
    )
}

/// 创建 Edit 控件类型条件
fn create_edit_condition(
    uia: &IUIAutomation,
) -> Option<IUIAutomationCondition> {
    unsafe {
        uia.CreatePropertyCondition(
            windows::Win32::UI::Accessibility::UIA_ControlTypePropertyId,
            &(UIA_EditControlTypeId.0 as i32).into(),
        )
        .ok()
    }
}

/// 在元素子树中查找 URL 编辑框
///
/// 遍历所有 Edit 控件，优先选择名称包含"地址"或"Address"的控件。
fn find_url_edit_element(
    root: &IUIAutomationElement,
    condition: &IUIAutomationCondition,
) -> Option<IUIAutomationElement> {
    unsafe {
        let elements = root
            .FindAll(TreeScope_Subtree, condition)
            .ok()?;

        let count = elements.Length().ok()?;
        if count == 0 {
            return None;
        }

        // 优先查找名称包含"地址"或"Address"的控件
        for i in 0..count {
            if let Ok(el) = elements.GetElement(i) {
                if let Ok(name) = el.CurrentName() {
                    let name_str = name.to_string();
                    if name_str.contains("地址")
                        || name_str.to_lowercase().contains("address")
                    {
                        return Some(el);
                    }
                }
            }
        }

        // 没有找到名称匹配的，取第一个 Edit 控件
        elements.GetElement(0).ok()
    }
}

/// 从 UI Automation 元素中读取 Value 属性
///
/// 使用 `IUIAutomationValuePattern` 获取文本值。
fn read_value_pattern(
    element: &IUIAutomationElement,
) -> Option<String> {
    unsafe {
        let pattern = element
            .GetCurrentPattern(UIA_ValuePatternId)
            .ok()?;

        let value_pattern: windows::Win32::UI::Accessibility::IUIAutomationValuePattern =
            pattern.cast().ok()?;

        let value = value_pattern.CurrentValue().ok()?;
        let url = value.to_string();

        // 验证是否为有效 URL
        if url.is_empty()
            || (!url.starts_with("http") && !url.starts_with("ftp"))
        {
            return None;
        }

        Some(url)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn chromium浏览器识别() {
        assert!(is_chromium_browser("chrome"));
        assert!(is_chromium_browser("msedge"));
        assert!(is_chromium_browser("brave"));
        assert!(is_chromium_browser("opera"));
        assert!(!is_chromium_browser("firefox"));
        assert!(!is_chromium_browser("safari"));
    }

    #[test]
    fn 非浏览器不应识别为chromium() {
        assert!(!is_chromium_browser("code"));
        assert!(!is_chromium_browser("wechat"));
    }
}

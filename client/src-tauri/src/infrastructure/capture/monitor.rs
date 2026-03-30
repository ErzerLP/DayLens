//! # Windows 窗口监控
//!
//! 通过 Win32 API 获取当前前台窗口信息。

use std::ffi::OsString;
use std::os::windows::ffi::OsStringExt;
use std::path::Path;

use windows::core::PWSTR;
use windows::Win32::Foundation::{CloseHandle, HWND, MAX_PATH};
use windows::Win32::System::Threading::{
    OpenProcess, QueryFullProcessImageNameW, PROCESS_NAME_FORMAT,
    PROCESS_QUERY_LIMITED_INFORMATION,
};
use windows::Win32::UI::WindowsAndMessaging::{
    GetForegroundWindow, GetWindowTextW, GetWindowThreadProcessId,
    IsWindowVisible,
};

use crate::domain::activity::entity::ActiveWindow;
use crate::shared::error::{AppError, Result};

// ===== 常量 =====

/// 应被过滤的系统窗口类列表
const SYSTEM_APP_NAMES: &[&str] = &[
    "explorer",         // 桌面
    "shellexperiencehost",
    "searchhost",
    "startmenuexperiencehost",
    "textinputhost",
    "lockapp",
    "logonui",
    "applicationframehost",
    "systemsettings",
];

/// 浏览器应用名列表（小写）
const BROWSER_NAMES: &[&str] = &[
    "chrome",
    "msedge",
    "firefox",
    "brave",
    "opera",
    "vivaldi",
    "arc",
    "chromium",
    "safari",
];

// ===== WindowsMonitor =====

/// Windows 窗口监控器
pub struct WindowsMonitor;

impl WindowsMonitor {
    /// 创建新的 WindowsMonitor
    pub fn new() -> Self {
        Self
    }

    /// 获取当前前台窗口信息
    pub fn get_active_window(&self) -> Result<ActiveWindow> {
        unsafe {
            // 1. 获取前台窗口句柄
            let hwnd = GetForegroundWindow();
            if hwnd.0 as usize == 0 {
                return Err(AppError::Platform(
                    "无前台窗口".to_string(),
                ));
            }

            // 检查窗口是否可见
            if !IsWindowVisible(hwnd).as_bool() {
                return Err(AppError::Platform(
                    "前台窗口不可见".to_string(),
                ));
            }

            // 2. 获取窗口标题
            let window_title = get_window_title(hwnd);

            // 3. 获取进程 ID
            let mut pid: u32 = 0;
            GetWindowThreadProcessId(hwnd, Some(&mut pid));
            if pid == 0 {
                return Err(AppError::Platform(
                    "无法获取窗口进程ID".to_string(),
                ));
            }

            // 4. 获取可执行文件路径
            let executable_path = get_executable_path(pid)
                .unwrap_or_default();

            // 5. 归一化应用名
            let app_name = normalize_app_name(&executable_path);

            // 6. 系统窗口过滤
            if is_system_window(&app_name, &window_title) {
                return Err(AppError::Platform(format!(
                    "系统窗口已过滤: {app_name}",
                )));
            }

            let is_browser = is_browser(&app_name);

            Ok(ActiveWindow {
                app_name,
                window_title,
                executable_path,
                is_browser,
            })
        }
    }

    /// 获取浏览器地址栏 URL（stub — 阶段 9 将实现 Accessibility API 提取）
    pub fn get_browser_url(&self, _app_name: &str) -> Option<String> {
        // TODO: 使用 UI Automation API 读取浏览器地址栏
        None
    }

    /// 获取当前前台窗口 HWND
    pub fn get_foreground_hwnd() -> HWND {
        unsafe { GetForegroundWindow() }
    }
}

// ===== 辅助函数 =====

/// 获取窗口标题
///
/// 调用 `GetWindowTextW()`，返回 UTF-16 → UTF-8 字符串。
fn get_window_title(hwnd: HWND) -> String {
    unsafe {
        let mut buf = [0u16; 512];
        let len = GetWindowTextW(hwnd, &mut buf);
        if len == 0 {
            return String::new();
        }
        String::from_utf16_lossy(&buf[..len as usize])
    }
}

/// 获取进程的可执行文件路径
///
/// 调用 `OpenProcess()` + `QueryFullProcessImageNameW()`。
fn get_executable_path(pid: u32) -> Option<String> {
    unsafe {
        let handle = OpenProcess(
            PROCESS_QUERY_LIMITED_INFORMATION,
            false,
            pid,
        )
        .ok()?;

        let mut buf = [0u16; MAX_PATH as usize];
        let mut size = buf.len() as u32;

        let ok = QueryFullProcessImageNameW(
            handle,
            PROCESS_NAME_FORMAT(0),
            PWSTR(buf.as_mut_ptr()),
            &mut size,
        );

        let _ = CloseHandle(handle);

        if ok.is_ok() {
            let os_str = OsString::from_wide(&buf[..size as usize]);
            Some(os_str.to_string_lossy().to_string())
        } else {
            None
        }
    }
}

/// 归一化应用名
///
/// 从可执行路径提取文件名，去掉 `.exe` 后缀，
/// 并将常见变体映射到标准名称。
///
/// # 示例
/// - `C:\...\Code.exe` → `Code`
/// - `C:\...\msedge.exe` → `msedge`
/// - `C:\...\WindowsTerminal.exe` → `WindowsTerminal`
pub fn normalize_app_name(executable_path: &str) -> String {
    if executable_path.is_empty() {
        return "Unknown".to_string();
    }

    // 从路径提取文件名
    let file_name = Path::new(executable_path)
        .file_stem()
        .and_then(|s| s.to_str())
        .unwrap_or("Unknown");

    // 变体映射
    let name_lower = file_name.to_lowercase();
    match name_lower.as_str() {
        "code" | "code - insiders" => "Code".to_string(),
        "msedge" => "msedge".to_string(),
        "chrome" => "Chrome".to_string(),
        "firefox" => "firefox".to_string(),
        "brave" => "Brave".to_string(),
        "idea64" | "idea" => "IDEA".to_string(),
        "pycharm64" | "pycharm" => "PyCharm".to_string(),
        "webstorm64" | "webstorm" => "WebStorm".to_string(),
        "goland64" | "goland" => "GoLand".to_string(),
        "clion64" | "clion" => "CLion".to_string(),
        "rider64" | "rider" => "Rider".to_string(),
        "rustrover64" | "rustrover" => "RustRover".to_string(),
        "datagrip64" | "datagrip" => "DataGrip".to_string(),
        "notepad++" => "Notepad++".to_string(),
        "sublime_text" => "Sublime Text".to_string(),
        "windowsterminal" => "WindowsTerminal".to_string(),
        "wechat" | "weixin" => "WeChat".to_string(),
        "dingtalk" => "DingTalk".to_string(),
        "feishu" | "lark" => "Feishu".to_string(),
        "winword" => "Word".to_string(),
        "excel" => "Excel".to_string(),
        "powerpnt" => "PowerPoint".to_string(),
        "onenote" => "OneNote".to_string(),
        "devenv" => "Visual Studio".to_string(),
        "taskmgr" => "Task Manager".to_string(),
        _ => file_name.to_string(),
    }
}

/// 判断应用是否为浏览器
pub fn is_browser(app_name: &str) -> bool {
    let name_lower = app_name.to_lowercase();
    BROWSER_NAMES.iter().any(|b| name_lower == *b)
}

/// 判断是否为应被过滤的系统窗口
///
/// 过滤条件：
/// 1. 应用名在系统窗口列表中
/// 2. 窗口标题为空（非可交互窗口）
fn is_system_window(app_name: &str, window_title: &str) -> bool {
    let name_lower = app_name.to_lowercase();

    // 系统应用列表
    if SYSTEM_APP_NAMES.iter().any(|s| name_lower == *s) {
        return true;
    }

    // 无标题窗口（通常是系统后台窗口）
    if window_title.trim().is_empty() {
        return true;
    }

    false
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 应用名归一化_code() {
        assert_eq!(
            normalize_app_name(r"C:\Users\test\AppData\Local\Programs\Microsoft VS Code\Code.exe"),
            "Code",
        );
    }

    #[test]
    fn 应用名归一化_chrome() {
        assert_eq!(
            normalize_app_name(r"C:\Program Files\Google\Chrome\Application\chrome.exe"),
            "Chrome",
        );
    }

    #[test]
    fn 应用名归一化_idea() {
        assert_eq!(
            normalize_app_name(r"C:\JetBrains\IntelliJ IDEA\bin\idea64.exe"),
            "IDEA",
        );
    }

    #[test]
    fn 应用名归一化_wechat() {
        assert_eq!(
            normalize_app_name(r"C:\Program Files\WeChat\WeChat.exe"),
            "WeChat",
        );
    }

    #[test]
    fn 应用名归一化_未知应用保留原名() {
        assert_eq!(
            normalize_app_name(r"C:\SomeApp\MyTool.exe"),
            "MyTool",
        );
    }

    #[test]
    fn 应用名归一化_空路径() {
        assert_eq!(normalize_app_name(""), "Unknown");
    }

    #[test]
    fn 浏览器识别_chrome() {
        assert!(is_browser("Chrome"));
        assert!(is_browser("chrome"));
    }

    #[test]
    fn 浏览器识别_edge() {
        assert!(is_browser("msedge"));
    }

    #[test]
    fn 浏览器识别_非浏览器() {
        assert!(!is_browser("Code"));
        assert!(!is_browser("WeChat"));
    }

    #[test]
    fn 系统窗口过滤_explorer() {
        assert!(is_system_window("explorer", ""));
    }

    #[test]
    fn 系统窗口过滤_空标题() {
        assert!(is_system_window("SomeApp", ""));
    }

    #[test]
    fn 系统窗口过滤_正常窗口() {
        assert!(!is_system_window("Code", "main.rs — daylens"));
    }

    #[test]
    fn 系统窗口过滤_lockapp() {
        assert!(is_system_window("LockApp", "锁屏"));
    }
}

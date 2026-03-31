//! # OCR 文字识别服务
//!
//! 从截图中识别文字内容。
//!
//! # 降级策略
//! 1. Windows.Media.Ocr（WinRT API，Windows 10+ 内置）
//! 2. PaddleOCR（Python 脚本，需要 Python 环境）
//! 3. 空文本降级（两者都不可用时返回空字符串）
//!
//! # 语言支持
//! Windows OCR 支持多语言（需要安装语言包），中英文默认可用。

use std::path::Path;
use std::process::Command;

#[cfg(target_os = "windows")]
use std::os::windows::process::CommandExt;

use crate::shared::error::{AppError, Result};

/// Windows `CREATE_NO_WINDOW` 标志，防止子进程弹出控制台窗口
#[cfg(target_os = "windows")]
const CREATE_NO_WINDOW: u32 = 0x08000000;

// ===== 常量 =====

/// PaddleOCR Python 脚本名（预期位于 resources/ 目录）
const PADDLE_SCRIPT: &str = "paddle_ocr.py";

/// OCR 结果最大长度限制（字符数）
const MAX_OCR_TEXT_LENGTH: usize = 5000;

// ===== OcrService =====

/// OCR 文字识别服务
#[derive(Clone)]
pub struct OcrService {
    /// PaddleOCR Python 是否可用
    paddle_available: bool,
    /// Windows OCR 是否可用
    windows_ocr_available: bool,
}

impl OcrService {
    /// 创建新的 OCR 服务
    ///
    /// 自动检测可用的 OCR 引擎。
    pub fn new() -> Self {
        let windows_ocr_available = Self::check_windows_ocr();
        let paddle_available = Self::check_paddle_ocr();

        log::info!(
            "OCR 引擎可用性: Windows OCR={}, PaddleOCR={}",
            windows_ocr_available,
            paddle_available,
        );

        Self {
            paddle_available,
            windows_ocr_available,
        }
    }

    /// 识别图片中的文字
    ///
    /// # 降级链路
    /// 1. 尝试 Windows OCR
    /// 2. Windows OCR 失败 → 尝试 PaddleOCR
    /// 3. 两者都失败 → 返回空字符串
    ///
    /// # 参数
    /// - `image_path` — 图片文件路径（支持 JPEG/PNG）
    pub fn recognize(&self, image_path: &Path) -> Result<String> {
        // 检查文件是否存在
        if !image_path.exists() {
            return Err(AppError::Platform(format!(
                "OCR 图片不存在: {}",
                image_path.display(),
            )));
        }

        // 尝试 Windows OCR
        if self.windows_ocr_available {
            match self.recognize_windows_ocr(image_path) {
                Ok(text) if !text.is_empty() => {
                    return Ok(truncate_text(text));
                }
                Ok(_) => {
                    log::debug!("Windows OCR 返回空文本，尝试降级");
                }
                Err(e) => {
                    log::warn!("Windows OCR 失败: {e}，尝试降级");
                }
            }
        }

        // 尝试 PaddleOCR
        if self.paddle_available {
            match self.recognize_paddle(image_path) {
                Ok(text) if !text.is_empty() => {
                    return Ok(truncate_text(text));
                }
                Ok(_) => {
                    log::debug!("PaddleOCR 返回空文本");
                }
                Err(e) => {
                    log::warn!("PaddleOCR 失败: {e}");
                }
            }
        }

        // 所有引擎都不可用或失败，返回空文本
        Ok(String::new())
    }

    /// 检查 OCR 引擎是否可用
    pub fn is_available(&self) -> bool {
        self.windows_ocr_available || self.paddle_available
    }

    /// Windows.Media.Ocr — WinRT OCR API
    ///
    /// 使用 PowerShell 调用 WinRT OCR API（避免直接依赖 WinRT 异步接口）。
    /// 这是一种实用的折中方案，简化了 WinRT 异步回调的处理。
    fn recognize_windows_ocr(
        &self,
        image_path: &Path,
    ) -> Result<String> {
        let path_str = image_path
            .to_str()
            .ok_or_else(|| AppError::Platform(
                "图片路径包含非法字符".to_string(),
            ))?;

        // 通过 PowerShell 调用 Windows OCR
        let script = format!(
            r#"
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$null = [Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType = WindowsRuntime]
$null = [Windows.Graphics.Imaging.BitmapDecoder, Windows.Foundation, ContentType = WindowsRuntime]

$path = '{}'
$stream = [System.IO.File]::OpenRead($path)
$randomAccessStream = [System.IO.WindowsRuntimeStreamExtensions]::AsRandomAccessStream($stream)

$decoder = [Windows.Graphics.Imaging.BitmapDecoder]::CreateAsync($randomAccessStream)
$decoder = $decoder.GetAwaiter().GetResult()

$bitmap = $decoder.GetSoftwareBitmapAsync()
$bitmap = $bitmap.GetAwaiter().GetResult()

$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
if ($engine -eq $null) {{
    $engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromLanguage(
        [Windows.Globalization.Language]::new('zh-Hans-CN')
    )
}}

if ($engine -ne $null) {{
    $result = $engine.RecognizeAsync($bitmap)
    $result = $result.GetAwaiter().GetResult()
    Write-Output $result.Text
}}

$stream.Dispose()
"#,
            path_str.replace('\\', "\\\\").replace('\'', "''"),
        );

        let mut cmd = Command::new("powershell");
        cmd.args(["-NoProfile", "-NonInteractive", "-Command", &script]);
        #[cfg(target_os = "windows")]
        cmd.creation_flags(CREATE_NO_WINDOW);
        let output = cmd
            .output()
            .map_err(|e| AppError::Platform(format!(
                "PowerShell 执行失败: {e}",
            )))?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(AppError::Platform(format!(
                "Windows OCR 执行失败: {stderr}",
            )));
        }

        let text = String::from_utf8_lossy(&output.stdout)
            .trim()
            .to_string();

        Ok(text)
    }

    /// PaddleOCR — Python 脚本调用
    ///
    /// 调用外部 Python 脚本执行 PaddleOCR 识别。
    fn recognize_paddle(
        &self,
        image_path: &Path,
    ) -> Result<String> {
        let path_str = image_path
            .to_str()
            .ok_or_else(|| AppError::Platform(
                "图片路径包含非法字符".to_string(),
            ))?;

        let mut cmd = Command::new("python");
        cmd.args([PADDLE_SCRIPT, path_str]);
        #[cfg(target_os = "windows")]
        cmd.creation_flags(CREATE_NO_WINDOW);
        let output = cmd
            .output()
            .map_err(|e| AppError::Platform(format!(
                "PaddleOCR 执行失败: {e}",
            )))?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(AppError::Platform(format!(
                "PaddleOCR 返回错误: {stderr}",
            )));
        }

        let text = String::from_utf8_lossy(&output.stdout)
            .trim()
            .to_string();

        Ok(text)
    }

    /// 检查 Windows OCR 是否可用
    ///
    /// 尝试加载 OcrEngine 类，不实际执行识别。
    fn check_windows_ocr() -> bool {
        // Windows 10 1809+ 内置 OCR，通过快速测试确认
        let mut cmd = Command::new("powershell");
        cmd.args([
                "-NoProfile",
                "-NonInteractive",
                "-Command",
                "[Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType = WindowsRuntime] | Out-Null; Write-Output 'ok'",
            ]);
        #[cfg(target_os = "windows")]
        cmd.creation_flags(CREATE_NO_WINDOW);
        let output = cmd.output();

        match output {
            Ok(o) => {
                o.status.success()
                    && String::from_utf8_lossy(&o.stdout)
                        .trim()
                        .contains("ok")
            }
            Err(_) => false,
        }
    }

    /// 检查 PaddleOCR 是否可用
    ///
    /// 尝试执行 `python -c "import paddleocr"`。
    fn check_paddle_ocr() -> bool {
        let mut cmd = Command::new("python");
        cmd.args(["-c", "import paddleocr; print('ok')"]);
        #[cfg(target_os = "windows")]
        cmd.creation_flags(CREATE_NO_WINDOW);
        let output = cmd.output();

        match output {
            Ok(o) => {
                o.status.success()
                    && String::from_utf8_lossy(&o.stdout)
                        .trim()
                        .contains("ok")
            }
            Err(_) => false,
        }
    }
}

// ===== 辅助函数 =====

/// 截断过长的 OCR 文本
fn truncate_text(text: String) -> String {
    if text.len() > MAX_OCR_TEXT_LENGTH {
        let mut truncated = text;
        truncated.truncate(MAX_OCR_TEXT_LENGTH);
        truncated.push_str("...[截断]");
        truncated
    } else {
        text
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 文本截断_短文本不截断() {
        let text = "Hello, World!".to_string();
        let result = truncate_text(text.clone());
        assert_eq!(result, text);
    }

    #[test]
    fn 文本截断_超长文本应截断() {
        let text = "a".repeat(6000);
        let result = truncate_text(text);
        assert!(result.len() < 6000);
        assert!(result.ends_with("...[截断]"));
    }

    #[test]
    fn ocr服务创建不应panic() {
        // 即使 OCR 引擎不可用，创建也不应失败
        let service = OcrService {
            paddle_available: false,
            windows_ocr_available: false,
        };
        assert!(!service.is_available());
    }

    #[test]
    fn 不存在的文件应返回错误() {
        let service = OcrService {
            paddle_available: false,
            windows_ocr_available: true,
        };
        let result = service.recognize(Path::new("nonexistent.jpg"));
        assert!(result.is_err());
    }

    #[test]
    fn 所有引擎不可用时应返回空文本() {
        let service = OcrService {
            paddle_available: false,
            windows_ocr_available: false,
        };
        // 创建一个临时文件以通过文件存在检查
        let temp_dir = std::env::temp_dir();
        let temp_file = temp_dir.join("test_ocr_empty.txt");
        std::fs::write(&temp_file, "test").unwrap();

        let result = service.recognize(&temp_file);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), "");

        let _ = std::fs::remove_file(&temp_file);
    }
}

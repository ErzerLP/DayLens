//! # 空闲检测器
//!
//! 通过 `GetLastInputInfo` Win32 API 检测用户是否处于空闲状态。
//! 如果距离最后一次键盘/鼠标输入超过阈值，则视为空闲。

use std::sync::atomic::{AtomicU32, Ordering};

use windows::Win32::UI::Input::KeyboardAndMouse::{
    GetLastInputInfo, LASTINPUTINFO,
};

// ===== InputIdleDetector =====

/// 输入空闲检测器
pub struct InputIdleDetector {
    /// 空闲阈值（毫秒）
    idle_threshold_ms: AtomicU32,
}

impl InputIdleDetector {
    /// 创建空闲检测器
    ///
    /// # 参数
    /// - `idle_timeout_minutes` — 空闲超时（分钟）
    pub fn new(idle_timeout_minutes: u32) -> Self {
        Self {
            idle_threshold_ms: AtomicU32::new(
                idle_timeout_minutes * 60 * 1000,
            ),
        }
    }

    /// 检查用户是否处于空闲状态
    pub fn is_idle(&self) -> bool {
        let idle_ms = self.get_idle_duration_ms();
        let threshold = self.idle_threshold_ms.load(Ordering::Relaxed);
        idle_ms >= threshold
    }

    /// 获取空闲时长（毫秒）
    pub fn get_idle_duration_ms(&self) -> u32 {
        unsafe {
            let mut info = LASTINPUTINFO {
                cbSize: std::mem::size_of::<LASTINPUTINFO>() as u32,
                dwTime: 0,
            };

            if GetLastInputInfo(&mut info).as_bool() {
                // 使用 kernel32 GetTickCount 获取系统运行时间
                // 由于我们没有直接绑定 GetTickCount，
                // 用 winapi-style 的 extern 调用
                extern "system" {
                    fn GetTickCount() -> u32;
                }
                let tick = GetTickCount();
                tick.wrapping_sub(info.dwTime)
            } else {
                0
            }
        }
    }

    /// 更新空闲阈值
    pub fn set_threshold_minutes(&self, minutes: u32) {
        self.idle_threshold_ms
            .store(minutes * 60 * 1000, Ordering::Relaxed);
    }

    /// 重置检测器（预留扩展）
    pub fn reset(&self) {
        // GetLastInputInfo 由系统自动重置，无需手动操作
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 创建空闲检测器() {
        let detector = InputIdleDetector::new(3);
        assert_eq!(
            detector.idle_threshold_ms.load(Ordering::Relaxed),
            180_000,
        );
    }

    #[test]
    fn 更新阈值() {
        let detector = InputIdleDetector::new(3);
        detector.set_threshold_minutes(5);
        assert_eq!(
            detector.idle_threshold_ms.load(Ordering::Relaxed),
            300_000,
        );
    }

    #[test]
    fn 获取空闲时长不应panic() {
        let detector = InputIdleDetector::new(3);
        let idle = detector.get_idle_duration_ms();
        // 在 CI 或测试环境中，空闲时长应该很短
        assert!(idle < 600_000); // < 10分钟
    }
}

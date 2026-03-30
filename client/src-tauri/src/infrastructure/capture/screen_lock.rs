//! # 锁屏检测器
//!
//! 通过 WTS Session API 检测 Windows 屏幕是否已锁定。
//! 锁屏期间应暂停采集，避免截取锁屏界面。

use std::sync::atomic::{AtomicBool, Ordering};

// ===== WtsLockDetector =====

/// WTS 锁屏检测器
pub struct WtsLockDetector {
    /// 当前是否锁屏
    locked: AtomicBool,
}

impl WtsLockDetector {
    /// 创建锁屏检测器
    pub fn new() -> Self {
        Self {
            locked: AtomicBool::new(false),
        }
    }

    /// 检查屏幕是否已锁定
    pub fn is_locked(&self) -> bool {
        self.locked.load(Ordering::Relaxed)
    }

    /// 标记屏幕已锁定
    ///
    /// 由 WTS 事件回调调用。
    pub fn set_locked(&self) {
        self.locked.store(true, Ordering::Relaxed);
        log::info!("屏幕已锁定，采集暂停");
    }

    /// 标记屏幕已解锁
    ///
    /// 由 WTS 事件回调调用。
    pub fn set_unlocked(&self) {
        self.locked.store(false, Ordering::Relaxed);
        log::info!("屏幕已解锁，采集恢复");
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 初始状态应为未锁定() {
        let detector = WtsLockDetector::new();
        assert!(!detector.is_locked());
    }

    #[test]
    fn 锁定和解锁状态切换() {
        let detector = WtsLockDetector::new();

        detector.set_locked();
        assert!(detector.is_locked());

        detector.set_unlocked();
        assert!(!detector.is_locked());
    }
}

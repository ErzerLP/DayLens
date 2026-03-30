//! # 配置管理用例
//!
//! 提供配置的读写和运行时更新。

use std::sync::RwLock;

use crate::domain::config::entity::AppConfig;
use crate::shared::error::Result;

/// 配置持久化端口
pub trait ConfigPersistence: Send + Sync {
    fn load(&self) -> Result<AppConfig>;
    fn save(&self, config: &AppConfig) -> Result<()>;
}

/// 配置管理器
pub struct ConfigManager {
    store: Box<dyn ConfigPersistence>,
    current: RwLock<AppConfig>,
}

impl ConfigManager {
    pub fn new(store: Box<dyn ConfigPersistence>) -> Result<Self> {
        let config = store.load()?;
        Ok(Self {
            store,
            current: RwLock::new(config),
        })
    }

    /// 获取当前配置快照
    pub fn get(&self) -> AppConfig {
        self.current.read().unwrap().clone()
    }

    /// 更新配置并持久化
    pub fn update<F>(&self, f: F) -> Result<()>
    where
        F: FnOnce(&mut AppConfig),
    {
        let mut config = self.current.write().unwrap();
        f(&mut config);
        self.store.save(&config)?;
        Ok(())
    }

    /// 重新加载配置
    pub fn reload(&self) -> Result<()> {
        let config = self.store.load()?;
        *self.current.write().unwrap() = config;
        Ok(())
    }
}

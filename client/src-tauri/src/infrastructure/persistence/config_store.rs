//! # 配置文件读写
//!
//! 将 `AppConfig` 序列化为 JSON 文件存储，启动时加载。
//! 文件路径：`%APPDATA%/daylens/config.json`

use std::fs;
use std::path::{Path, PathBuf};

use crate::domain::config::entity::AppConfig;
use crate::shared::error::{AppError, Result};

// ===== ConfigStore =====

/// 配置文件存储服务
///
/// 负责将 `AppConfig` 持久化到本地 JSON 文件。
pub struct ConfigStore {
    /// 配置文件路径
    config_path: PathBuf,
}

impl ConfigStore {
    /// 创建配置存储服务
    ///
    /// # 参数
    /// - `config_path` — 配置文件完整路径
    pub fn new(config_path: PathBuf) -> Self {
        Self { config_path }
    }

    /// 从默认位置创建
    ///
    /// 默认路径：`%APPDATA%/daylens/config.json`
    pub fn from_default() -> Result<Self> {
        let app_dir = dirs::config_dir()
            .ok_or_else(|| AppError::Config(
                "无法获取配置目录".to_string(),
            ))?
            .join("daylens");

        fs::create_dir_all(&app_dir)?;
        Ok(Self::new(app_dir.join("config.json")))
    }

    /// 加载配置
    ///
    /// 如果文件不存在，返回默认配置并保存。
    pub fn load(&self) -> Result<AppConfig> {
        if !self.config_path.exists() {
            let default = AppConfig::default();
            self.save(&default)?;
            return Ok(default);
        }

        let content = fs::read_to_string(&self.config_path)?;
        let config: AppConfig = serde_json::from_str(&content)
            .map_err(|e| AppError::Config(format!(
                "配置文件解析失败: {e}",
            )))?;

        Ok(config)
    }

    /// 保存配置
    pub fn save(&self, config: &AppConfig) -> Result<()> {
        // 确保目录存在
        if let Some(parent) = self.config_path.parent() {
            fs::create_dir_all(parent)?;
        }

        let json = serde_json::to_string_pretty(config)
            .map_err(|e| AppError::Config(format!(
                "配置序列化失败: {e}",
            )))?;

        fs::write(&self.config_path, json)?;
        Ok(())
    }

    /// 获取配置文件路径
    pub fn config_path(&self) -> &Path {
        &self.config_path
    }

    /// 检查配置文件是否存在
    pub fn exists(&self) -> bool {
        self.config_path.exists()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn temp_config_path(name: &str) -> PathBuf {
        std::env::temp_dir()
            .join("wr_test_config")
            .join(format!("{name}.json"))
    }

    #[test]
    fn 加载不存在的配置应返回默认值() {
        let path = temp_config_path("default_load");
        let _ = fs::remove_file(&path);

        let store = ConfigStore::new(path.clone());
        let config = store.load().expect("加载失败");

        assert_eq!(config.server.url, "http://localhost:8080");
        assert!(store.exists());

        // 清理
        let _ = fs::remove_file(&path);
    }

    #[test]
    fn 保存并重新加载应一致() {
        let path = temp_config_path("save_load");
        let _ = fs::remove_file(&path);

        let store = ConfigStore::new(path.clone());
        let mut config = AppConfig::default();
        config.server.url = "https://example.com".to_string();
        config.server.token = "test-token".to_string();

        store.save(&config).expect("保存失败");
        let loaded = store.load().expect("加载失败");

        assert_eq!(loaded.server.url, "https://example.com");
        assert_eq!(loaded.server.token, "test-token");

        // 清理
        let _ = fs::remove_file(&path);
    }

    #[test]
    fn 损坏的配置文件应返回错误() {
        let path = temp_config_path("corrupt");
        let _ = fs::create_dir_all(path.parent().unwrap());
        fs::write(&path, "not valid json").unwrap();

        let store = ConfigStore::new(path.clone());
        let result = store.load();

        assert!(result.is_err());

        let _ = fs::remove_file(&path);
    }
}

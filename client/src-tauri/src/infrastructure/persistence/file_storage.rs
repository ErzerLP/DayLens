//! # 截图文件管理
//!
//! 管理截图文件的存储和清理：
//! - 按日期目录组织（`screenshots/2026-03-29/`）
//! - 自动清理过期文件
//! - UUID 命名，防冲突

use std::fs;
use std::path::{Path, PathBuf};

use chrono::Local;
use uuid::Uuid;

use crate::shared::error::{AppError, Result};

// ===== FileStorage =====

/// 截图文件存储服务
///
/// 负责截图文件的组织、路径生成和过期清理。
pub struct FileStorage {
    /// 基础存储目录（如 `%APPDATA%/daylens/screenshots`）
    base_dir: PathBuf,
}

impl FileStorage {
    /// 创建新的文件存储服务
    ///
    /// # 参数
    /// - `base_dir` — 截图基础存储目录
    pub fn new(base_dir: PathBuf) -> Self {
        Self { base_dir }
    }

    /// 生成截图保存路径
    ///
    /// 格式：`{base_dir}/{date}/{uuid}.jpg`
    ///
    /// # 返回
    /// `(full_path, storage_key)`
    /// - `full_path` — 本地绝对路径
    /// - `storage_key` — 相对路径（上报给服务端）
    pub fn generate_screenshot_path(&self) -> Result<(PathBuf, String)> {
        let date = Local::now().format("%Y-%m-%d").to_string();
        let filename = format!("{}.jpg", Uuid::new_v4());
        let storage_key = format!("{date}/{filename}");

        let dir = self.base_dir.join(&date);
        fs::create_dir_all(&dir)?;

        let full_path = dir.join(&filename);
        Ok((full_path, storage_key))
    }

    /// 生成缩略图路径
    ///
    /// 格式：`{base_dir}/{date}/thumb_{uuid}.jpg`
    pub fn generate_thumbnail_path(
        &self,
        screenshot_path: &Path,
    ) -> PathBuf {
        let parent = screenshot_path
            .parent()
            .unwrap_or(&self.base_dir);
        let stem = screenshot_path
            .file_stem()
            .and_then(|s| s.to_str())
            .unwrap_or("unknown");

        parent.join(format!("thumb_{stem}.jpg"))
    }

    /// 清理指定日期之前的截图
    ///
    /// 删除 `{base_dir}/{date}/` 目录及其所有文件。
    ///
    /// # 参数
    /// - `before_date` — YYYY-MM-DD 格式，删除此日期之前的所有目录
    ///
    /// # 返回
    /// 删除的文件数量
    pub fn cleanup_before(&self, before_date: &str) -> Result<u64> {
        let mut deleted_count = 0u64;

        if !self.base_dir.exists() {
            return Ok(0);
        }

        let entries = fs::read_dir(&self.base_dir)
            .map_err(|e| AppError::Io(e))?;

        for entry in entries.flatten() {
            let name = entry
                .file_name()
                .to_string_lossy()
                .to_string();

            // 只处理日期格式的目录（YYYY-MM-DD）
            if name.len() == 10
                && name.chars().nth(4) == Some('-')
                && name.as_str() < before_date
                && entry.path().is_dir()
            {
                // 统计文件数
                if let Ok(dir_entries) = fs::read_dir(entry.path()) {
                    deleted_count += dir_entries
                        .filter_map(|e| e.ok())
                        .filter(|e| e.path().is_file())
                        .count() as u64;
                }

                // 删除整个日期目录
                fs::remove_dir_all(entry.path())
                    .map_err(|e| AppError::Io(e))?;

                log::info!(
                    "已清理截图目录: {}",
                    entry.path().display(),
                );
            }
        }

        Ok(deleted_count)
    }

    /// 获取存储统计信息
    ///
    /// # 返回
    /// `(file_count, total_size_bytes)`
    pub fn get_storage_stats(&self) -> Result<(u64, u64)> {
        let mut file_count = 0u64;
        let mut total_size = 0u64;

        if !self.base_dir.exists() {
            return Ok((0, 0));
        }

        // 遍历所有日期目录
        for entry in fs::read_dir(&self.base_dir)
            .map_err(|e| AppError::Io(e))?
            .flatten()
        {
            if entry.path().is_dir() {
                if let Ok(files) = fs::read_dir(entry.path()) {
                    for file in files.flatten() {
                        if file.path().is_file() {
                            file_count += 1;
                            if let Ok(meta) = file.metadata() {
                                total_size += meta.len();
                            }
                        }
                    }
                }
            }
        }

        Ok((file_count, total_size))
    }

    /// 获取基础目录
    pub fn base_dir(&self) -> &Path {
        &self.base_dir
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 生成截图路径格式正确() {
        let temp = std::env::temp_dir().join("wr_test_fs_path");
        let storage = FileStorage::new(temp.clone());

        let (full_path, key) = storage
            .generate_screenshot_path()
            .expect("生成路径失败");

        // 路径应在基础目录下
        assert!(full_path.starts_with(&temp));
        // key 应为 YYYY-MM-DD/uuid.jpg 格式
        assert!(key.ends_with(".jpg"));
        assert!(key.contains('/'));
        // 日期部分长10个字符
        let date_part = key.split('/').next().unwrap();
        assert_eq!(date_part.len(), 10);

        // 清理
        let _ = fs::remove_dir_all(&temp);
    }

    #[test]
    fn 缩略图路径应包含thumb前缀() {
        let temp = std::env::temp_dir().join("wr_test_fs_thumb");
        let storage = FileStorage::new(temp.clone());

        let screenshot = temp.join("2026-03-29").join("abc123.jpg");
        let thumb = storage.generate_thumbnail_path(&screenshot);

        assert!(
            thumb.to_string_lossy().contains("thumb_abc123"),
        );
    }

    #[test]
    fn 存储统计_空目录() {
        let temp = std::env::temp_dir().join("wr_test_fs_stats_empty");
        let _ = fs::remove_dir_all(&temp);
        let storage = FileStorage::new(temp);

        let (count, size) = storage
            .get_storage_stats()
            .expect("统计失败");
        assert_eq!(count, 0);
        assert_eq!(size, 0);
    }

    #[test]
    fn 清理_空目录不应报错() {
        let temp = std::env::temp_dir()
            .join("wr_test_fs_cleanup_empty");
        let _ = fs::remove_dir_all(&temp);
        let storage = FileStorage::new(temp);

        let count = storage
            .cleanup_before("2026-03-29")
            .expect("清理失败");
        assert_eq!(count, 0);
    }

    #[test]
    fn 清理_应删除过期目录() {
        let temp = std::env::temp_dir()
            .join("wr_test_fs_cleanup");
        let _ = fs::remove_dir_all(&temp);

        // 创建模拟目录
        let old_dir = temp.join("2026-03-01");
        let new_dir = temp.join("2026-03-29");
        fs::create_dir_all(&old_dir).unwrap();
        fs::create_dir_all(&new_dir).unwrap();

        // 在旧目录中创建文件
        fs::write(old_dir.join("test1.jpg"), "fake").unwrap();
        fs::write(old_dir.join("test2.jpg"), "fake").unwrap();
        // 在新目录中创建文件
        fs::write(new_dir.join("test3.jpg"), "fake").unwrap();

        let storage = FileStorage::new(temp.clone());
        let deleted = storage
            .cleanup_before("2026-03-15")
            .expect("清理失败");

        assert_eq!(deleted, 2);
        assert!(!old_dir.exists());
        assert!(new_dir.exists());

        let _ = fs::remove_dir_all(&temp);
    }
}

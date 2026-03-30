//! # 领域层
//!
//! 包含所有业务实体、值对象和纯逻辑的领域服务。
//! 本层不依赖任何外部框架（无 reqwest / rusqlite / tokio / tauri / windows）。

pub mod activity;
pub mod config;
pub mod sync;

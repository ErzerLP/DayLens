//! # 接口层
//!
//! Tauri 命令、事件定义和应用状态管理。
//! 命令文件仅通过 trait 对象调用应用层服务，禁止引用 infrastructure。

pub mod commands;
pub mod events;
pub mod state;

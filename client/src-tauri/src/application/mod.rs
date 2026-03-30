//! # 应用层
//!
//! 包含用例编排和端口（trait）定义。
//! 通过 trait 对象与基础设施层解耦，绝不引用具体实现类型。

pub mod ports;
pub mod capture;
pub mod query;
pub mod sync;
pub mod config;

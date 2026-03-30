//! # 截屏服务
//!
//! 提供屏幕截取和缩略图生成功能。
//! 使用 GDI BitBlt 作为主要截屏方式（兼容性好）。
//!
//! # 降级策略
//! 1. 主方案：GDI BitBlt（保底，所有 Windows 版本可用）
//! 2. 缩略图：`image` crate 缩放到指定宽度
//!
//! # 线程安全
//! GDI 操作必须在同一线程完成（DC 句柄不能跨线程）。

use std::path::Path;

use image::codecs::jpeg::JpegEncoder;
use image::{GenericImageView, ImageReader};

use windows::Win32::Foundation::HWND;
use windows::Win32::Graphics::Gdi::{
    BitBlt, CreateCompatibleBitmap, CreateCompatibleDC, DeleteDC,
    DeleteObject, GetDC, GetDIBits, ReleaseDC, SelectObject, BITMAPINFO,
    BITMAPINFOHEADER, BI_RGB, DIB_RGB_COLORS, SRCCOPY,
};
use windows::Win32::UI::WindowsAndMessaging::{
    GetSystemMetrics, SM_CXSCREEN, SM_CYSCREEN,
};

use crate::shared::error::{AppError, Result};

// ===== 常量 =====

/// 默认 JPEG 质量（0-100）
const JPEG_QUALITY: u8 = 80;

/// 默认缩略图宽度（像素）
const DEFAULT_THUMBNAIL_WIDTH: u32 = 360;

// ===== ScreenshotService =====

/// GDI 截屏服务
pub struct ScreenshotService;

impl ScreenshotService {
    /// 创建新的截屏服务
    pub fn new() -> Self {
        Self
    }

    /// 截取当前屏幕并保存为 JPEG
    ///
    /// 使用 GDI BitBlt 截取整个桌面。
    ///
    /// # 参数
    /// - `save_path` — JPEG 保存路径
    pub fn capture(&self, save_path: &Path) -> Result<()> {
        self.capture_gdi(save_path)
    }

    /// GDI BitBlt 截屏
    fn capture_gdi(&self, save_path: &Path) -> Result<()> {
        unsafe {
            // 获取屏幕尺寸
            let width = GetSystemMetrics(SM_CXSCREEN);
            let height = GetSystemMetrics(SM_CYSCREEN);
            if width <= 0 || height <= 0 {
                return Err(AppError::Platform(
                    "无法获取屏幕尺寸".to_string(),
                ));
            }

            // 获取桌面 DC
            let screen_dc = GetDC(HWND(std::ptr::null_mut()));
            if screen_dc.is_invalid() {
                return Err(AppError::Platform(
                    "无法获取桌面DC".to_string(),
                ));
            }

            // 创建兼容 DC 和位图
            let mem_dc = CreateCompatibleDC(screen_dc);
            let bitmap = CreateCompatibleBitmap(
                screen_dc,
                width,
                height,
            );
            let old_bitmap = SelectObject(mem_dc, bitmap);

            // BitBlt 复制屏幕内容
            let blt_result = BitBlt(
                mem_dc,
                0,
                0,
                width,
                height,
                screen_dc,
                0,
                0,
                SRCCOPY,
            );

            if blt_result.is_err() {
                // 清理资源
                SelectObject(mem_dc, old_bitmap);
                let _ = DeleteObject(bitmap);
                let _ = DeleteDC(mem_dc);
                ReleaseDC(HWND(std::ptr::null_mut()), screen_dc);
                return Err(AppError::Platform(
                    "BitBlt 截屏失败".to_string(),
                ));
            }

            // 准备 BITMAPINFO 结构
            let mut bmi = BITMAPINFO {
                bmiHeader: BITMAPINFOHEADER {
                    biSize: std::mem::size_of::<BITMAPINFOHEADER>()
                        as u32,
                    biWidth: width,
                    biHeight: -height, // 负值表示 top-down
                    biPlanes: 1,
                    biBitCount: 32,
                    biCompression: BI_RGB.0 as u32,
                    ..Default::default()
                },
                ..Default::default()
            };

            // 分配像素缓冲区
            let buf_size = (width * height * 4) as usize;
            let mut pixels = vec![0u8; buf_size];

            // 读取像素数据
            let scan_lines = GetDIBits(
                mem_dc,
                bitmap,
                0,
                height as u32,
                Some(pixels.as_mut_ptr().cast()),
                &mut bmi,
                DIB_RGB_COLORS,
            );

            // 清理 GDI 资源
            SelectObject(mem_dc, old_bitmap);
            let _ = DeleteObject(bitmap);
            let _ = DeleteDC(mem_dc);
            ReleaseDC(HWND(std::ptr::null_mut()), screen_dc);

            if scan_lines == 0 {
                return Err(AppError::Platform(
                    "GetDIBits 读取像素失败".to_string(),
                ));
            }

            // BGRA → RGBA 转换
            for chunk in pixels.chunks_exact_mut(4) {
                chunk.swap(0, 2); // B ↔ R
            }

            // 保存为 JPEG
            save_jpeg(
                &pixels,
                width as u32,
                height as u32,
                save_path,
            )?;

            Ok(())
        }
    }

    /// 生成缩略图（按宽度等比缩放）
    pub fn generate_thumbnail(
        &self,
        source: &Path,
        target: &Path,
        width: u32,
    ) -> Result<()> {
        let img = ImageReader::open(source)
            .map_err(|e| AppError::Io(e.into()))?
            .decode()
            .map_err(|e| AppError::Platform(format!(
                "图片解码失败: {e}",
            )))?;

        let (orig_w, orig_h) = img.dimensions();
        if orig_w == 0 {
            return Err(AppError::Platform(
                "原图宽度为零".to_string(),
            ));
        }

        let target_width = width.min(orig_w);
        let target_height = (orig_h as f64
            * target_width as f64
            / orig_w as f64) as u32;

        let thumbnail = img.resize_exact(
            target_width,
            target_height.max(1),
            image::imageops::FilterType::Lanczos3,
        );

        // 确保目标目录存在
        if let Some(parent) = target.parent() {
            std::fs::create_dir_all(parent)?;
        }

        thumbnail
            .save(target)
            .map_err(|e| AppError::Platform(format!(
                "缩略图保存失败: {e}",
            )))?;

        Ok(())
    }

    /// 获取默认缩略图宽度
    pub fn default_thumbnail_width() -> u32 {
        DEFAULT_THUMBNAIL_WIDTH
    }
}

// ===== 辅助函数 =====

/// 将 RGBA 像素数据保存为 JPEG 文件
fn save_jpeg(
    pixels: &[u8],
    width: u32,
    height: u32,
    path: &Path,
) -> Result<()> {
    // 确保目标目录存在
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }

    // RGBA → RGB（JPEG 不支持 alpha 通道）
    let rgb_pixels: Vec<u8> = pixels
        .chunks_exact(4)
        .flat_map(|rgba| &rgba[..3])
        .copied()
        .collect();

    let file = std::fs::File::create(path)?;
    let mut encoder = JpegEncoder::new_with_quality(
        std::io::BufWriter::new(file),
        JPEG_QUALITY,
    );

    encoder
        .encode(&rgb_pixels, width, height, image::ExtendedColorType::Rgb8)
        .map_err(|e| AppError::Platform(format!(
            "JPEG 编码失败: {e}",
        )))?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 默认缩略图宽度应为360() {
        assert_eq!(
            ScreenshotService::default_thumbnail_width(),
            360,
        );
    }

    #[test]
    fn jpeg质量常量应合理() {
        assert!(JPEG_QUALITY > 50);
        assert!(JPEG_QUALITY <= 100);
    }

    #[test]
    fn rgba转rgb应正确() {
        let rgba = [255u8, 128, 64, 255, 0, 0, 0, 128];
        let rgb: Vec<u8> = rgba
            .chunks_exact(4)
            .flat_map(|c| &c[..3])
            .copied()
            .collect();
        assert_eq!(rgb, vec![255, 128, 64, 0, 0, 0]);
    }
}

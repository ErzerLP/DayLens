// 格式化工具函数

/** 分钟 → "Xh Ym" */
export function formatDuration(minutes: number): string {
  if (minutes < 1) return "< 1m";
  const h = Math.floor(minutes / 60);
  const m = Math.round(minutes % 60);
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

/** 秒 → "Xh Ym Zs" */
export function formatSeconds(secs: number): string {
  return formatDuration(secs / 60);
}

/** 字节 → "X MB" */
export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

/** ISO 时间 → "HH:mm" */
export function formatTime(iso: string): string {
  try {
    const d = new Date(iso);
    return d.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit", hour12: false });
  } catch {
    return iso;
  }
}

/** ISO 日期 → "YYYY-MM-DD" */
export function formatDate(iso: string): string {
  try {
    const d = new Date(iso);
    return d.toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" });
  } catch {
    return iso;
  }
}

/** 获取今天 YYYY-MM-DD */
export function today(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

/** 分类 → 中文 */
const CATEGORY_LABELS: Record<string, string> = {
  coding: "编码开发",
  browser: "网页浏览",
  communication: "即时通讯",
  document: "文档编辑",
  design: "设计工具",
  terminal: "终端命令",
  media: "媒体播放",
  system: "系统工具",
  gaming: "游戏娱乐",
  other: "其他",
};

export function categoryLabel(cat: string): string {
  return CATEGORY_LABELS[cat] ?? cat;
}

/** 分类 → HSL 颜色 */
const CATEGORY_COLORS: Record<string, string> = {
  coding: "hsl(210, 90%, 60%)",
  browser: "hsl(38, 92%, 55%)",
  communication: "hsl(145, 65%, 48%)",
  document: "hsl(260, 70%, 60%)",
  design: "hsl(330, 70%, 55%)",
  terminal: "hsl(180, 50%, 50%)",
  media: "hsl(0, 75%, 55%)",
  system: "hsl(220, 14%, 50%)",
  gaming: "hsl(290, 60%, 55%)",
  other: "hsl(0, 0%, 50%)",
};

export function categoryColor(cat: string): string {
  return CATEGORY_COLORS[cat] ?? "hsl(0, 0%, 50%)";
}

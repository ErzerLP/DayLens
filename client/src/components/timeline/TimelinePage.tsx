import { useEffect, useState, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Calendar, Image, X, ChevronLeft, ChevronRight } from "lucide-react";
import { useInvoke } from "../../hooks/useInvoke";
import { CMD } from "../../utils/api";
import { formatTime, formatSeconds, categoryLabel, categoryColor, today } from "../../utils/format";
import type { Activity, TimelineResponse } from "../../types";
import "./Timeline.css";

const PAGE_SIZE = 30;

export default function TimelinePage() {
  const [date, setDate] = useState(today());
  const [activities, setActivities] = useState<Activity[]>([]);
  const [hasMore, setHasMore] = useState(true);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const timeline = useInvoke<TimelineResponse>(CMD.GET_TIMELINE);

  const loadPage = useCallback(
    async (reset = false) => {
      const offset = reset ? 0 : activities.length;
      const res = await timeline.execute({
        date,
        limit: PAGE_SIZE,
        offset,
        app: null,
        category: null,
      });
      if (res) {
        setActivities((prev) => (reset ? res.items : [...prev, ...res.items]));
        setHasMore(res.hasMore);
      }
    },
    [date, activities.length, timeline],
  );

  useEffect(() => {
    setActivities([]);
    setSelectedId(null);
    loadPage(true);
  }, [date]); // eslint-disable-line react-hooks/exhaustive-deps

  const selected = activities.find((a) => a.id === selectedId) ?? null;

  const prevDay = () => {
    const d = new Date(date);
    d.setDate(d.getDate() - 1);
    setDate(d.toISOString().slice(0, 10));
  };

  const nextDay = () => {
    const d = new Date(date);
    d.setDate(d.getDate() + 1);
    const t = today();
    const next = d.toISOString().slice(0, 10);
    if (next <= t) setDate(next);
  };

  return (
    <div className="timeline-page">
      {/* 日期选择器 */}
      <div className="timeline-page__date-bar">
        <button className="icon-btn" onClick={prevDay}>
          <ChevronLeft size={18} />
        </button>
        <div className="timeline-page__date">
          <Calendar size={16} />
          <input
            type="date"
            value={date}
            max={today()}
            onChange={(e) => setDate(e.target.value)}
            className="date-input"
          />
        </div>
        <button className="icon-btn" onClick={nextDay} disabled={date >= today()}>
          <ChevronRight size={18} />
        </button>
        <span className="timeline-page__count">
          {activities.length} 条记录
        </span>
      </div>

      <div className="timeline-page__body">
        {/* 左侧列表 */}
        <div className="timeline-list">
          {activities.map((act) => (
            <motion.div
              key={act.id}
              className={`timeline-item ${selectedId === act.id ? "timeline-item--selected" : ""}`}
              onClick={() => setSelectedId(act.id)}
              initial={{ opacity: 0, x: -10 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.2 }}
            >
              <div className="timeline-item__time">{formatTime(act.startedAt)}</div>
              <div className="timeline-item__content">
                <div className="timeline-item__app">
                  <span
                    className="timeline-item__category-dot"
                    style={{ background: categoryColor(act.category) }}
                  />
                  {act.appName}
                  {act.screenshotPath && (
                    <Image size={12} className="timeline-item__screenshot-icon" />
                  )}
                </div>
                <div className="timeline-item__title">{act.windowTitle}</div>
              </div>
              <div className="timeline-item__duration">{formatSeconds(act.durationSecs)}</div>
            </motion.div>
          ))}

          {hasMore && (
            <button
              className="timeline-list__load-more"
              onClick={() => loadPage(false)}
              disabled={timeline.loading}
            >
              {timeline.loading ? "加载中…" : "加载更多"}
            </button>
          )}

          {!timeline.loading && activities.length === 0 && (
            <div className="timeline-list__empty">当天暂无活动记录</div>
          )}
        </div>

        {/* 右侧检查器 */}
        <AnimatePresence>
          {selected && (
            <motion.div
              className="timeline-inspector"
              initial={{ opacity: 0, x: 40 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: 40 }}
              transition={{ duration: 0.25 }}
            >
              <div className="timeline-inspector__header">
                <h3>{selected.appName}</h3>
                <button className="icon-btn" onClick={() => setSelectedId(null)}>
                  <X size={16} />
                </button>
              </div>

              {selected.screenshotPath && (
                <div className="timeline-inspector__screenshot">
                  <img
                    src={`asset://localhost/${selected.thumbnailPath ?? selected.screenshotPath}`}
                    alt="截图预览"
                  />
                </div>
              )}

              <div className="timeline-inspector__meta">
                <div className="meta-row">
                  <span className="meta-label">时间</span>
                  <span>{formatTime(selected.startedAt)} — {formatTime(selected.endedAt)}</span>
                </div>
                <div className="meta-row">
                  <span className="meta-label">时长</span>
                  <span>{formatSeconds(selected.durationSecs)}</span>
                </div>
                <div className="meta-row">
                  <span className="meta-label">分类</span>
                  <span
                    className="meta-badge"
                    style={{ background: categoryColor(selected.category) + "22", color: categoryColor(selected.category) }}
                  >
                    {categoryLabel(selected.category)}
                  </span>
                </div>
                <div className="meta-row">
                  <span className="meta-label">语义</span>
                  <span>{selected.semanticCategory}</span>
                </div>
              </div>

              <div className="timeline-inspector__title">
                <div className="meta-label">窗口标题</div>
                <div>{selected.windowTitle}</div>
              </div>

              {selected.ocrText && (
                <div className="timeline-inspector__ocr">
                  <div className="meta-label">OCR 文本</div>
                  <pre>{selected.ocrText}</pre>
                </div>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
}

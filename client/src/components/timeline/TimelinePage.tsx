import { useEffect, useState, useCallback, useMemo } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Calendar, Image, X, ChevronLeft, ChevronRight, ChevronDown } from "lucide-react";
import { useInvoke } from "../../hooks/useInvoke";
import { CMD } from "../../utils/api";
import { formatTime, formatSeconds, categoryLabel, categoryColor, today } from "../../utils/format";
import type { Activity, TimelineResponse } from "../../types";
import "./Timeline.css";

const PAGE_SIZE = 50;

// 连续相同应用的合并组
interface ActivityGroup {
  key: string;
  appName: string;
  category: string;
  startTime: number;
  endTime: number;
  totalDuration: number;
  activities: Activity[];
  hasScreenshot: boolean;
  mainTitle: string; // 最长使用的窗口标题
}

/** 将连续相同 appName 的活动合并 */
function mergeActivities(activities: Activity[]): ActivityGroup[] {
  if (activities.length === 0) return [];

  const groups: ActivityGroup[] = [];
  let current: ActivityGroup | null = null;

  for (const act of activities) {
    const ts = act.timestamp ?? new Date(act.startedAt).getTime() / 1000;
    const dur = act.duration ?? act.durationSecs ?? 30;

    if (current && current.appName === act.appName) {
      // 合入当前组
      current.activities.push(act);
      current.totalDuration += dur;
      current.endTime = ts + dur;
      if (act.screenshotPath) current.hasScreenshot = true;
    } else {
      // 新组
      current = {
        key: `group-${act.id}`,
        appName: act.appName,
        category: act.category,
        startTime: ts,
        endTime: ts + dur,
        totalDuration: dur,
        activities: [act],
        hasScreenshot: !!act.screenshotPath,
        mainTitle: act.windowTitle,
      };
      groups.push(current);
    }
  }

  // 计算每组最常出现的窗口标题
  for (const g of groups) {
    const titleMap = new Map<string, number>();
    for (const a of g.activities) {
      const d = a.duration ?? a.durationSecs ?? 30;
      titleMap.set(a.windowTitle, (titleMap.get(a.windowTitle) ?? 0) + d);
    }
    let maxDur = 0;
    for (const [title, dur] of titleMap) {
      if (dur > maxDur) {
        maxDur = dur;
        g.mainTitle = title;
      }
    }
  }

  return groups;
}

export default function TimelinePage() {
  const [date, setDate] = useState(today());
  const [activities, setActivities] = useState<Activity[]>([]);
  const [hasMore, setHasMore] = useState(true);
  const [selectedGroupKey, setSelectedGroupKey] = useState<string | null>(null);
  const [expandedGroup, setExpandedGroup] = useState<string | null>(null);
  const [selectedActivity, setSelectedActivity] = useState<Activity | null>(null);
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
    setSelectedGroupKey(null);
    setExpandedGroup(null);
    setSelectedActivity(null);
    loadPage(true);
  }, [date]); // eslint-disable-line react-hooks/exhaustive-deps

  const groups = useMemo(() => mergeActivities(activities), [activities]);

  const handleGroupClick = (group: ActivityGroup) => {
    setSelectedGroupKey(group.key);
    // 默认选中组内第一条有截图的，否则第一条
    const withScreenshot = group.activities.find((a) => a.screenshotPath);
    setSelectedActivity(withScreenshot ?? group.activities[0]);
  };

  const toggleExpand = (key: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setExpandedGroup(expandedGroup === key ? null : key);
  };

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

  const selected = selectedActivity;

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
          {activities.length} 条记录 · {groups.length} 个段落
        </span>
      </div>

      <div className="timeline-page__body">
        {/* 左侧列表 */}
        <div className="timeline-list">
          {groups.map((group) => (
            <div key={group.key}>
              <motion.div
                className={`timeline-group ${selectedGroupKey === group.key ? "timeline-group--selected" : ""}`}
                onClick={() => handleGroupClick(group)}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.2 }}
              >
                <div className="timeline-group__time">
                  <span>{formatTime(group.startTime)}</span>
                  <span className="timeline-group__time-end">{formatTime(group.endTime)}</span>
                </div>
                <div className="timeline-group__bar" style={{ background: categoryColor(group.category) }} />
                <div className="timeline-group__content">
                  <div className="timeline-group__app">
                    <span
                      className="timeline-item__category-dot"
                      style={{ background: categoryColor(group.category) }}
                    />
                    {group.appName}
                    {group.hasScreenshot && (
                      <Image size={12} className="timeline-item__screenshot-icon" />
                    )}
                    {group.activities.length > 1 && (
                      <span className="timeline-group__count">
                        ×{group.activities.length}
                      </span>
                    )}
                  </div>
                  <div className="timeline-item__title">{group.mainTitle}</div>
                </div>
                <div className="timeline-group__duration">
                  {formatSeconds(group.totalDuration)}
                </div>
                {group.activities.length > 1 && (
                  <button
                    className="timeline-group__expand"
                    onClick={(e) => toggleExpand(group.key, e)}
                    title="展开详情"
                  >
                    <ChevronDown
                      size={14}
                      style={{
                        transform: expandedGroup === group.key ? "rotate(180deg)" : "rotate(0)",
                        transition: "transform 0.2s",
                      }}
                    />
                  </button>
                )}
              </motion.div>

              {/* 展开详情 */}
              <AnimatePresence>
                {expandedGroup === group.key && (
                  <motion.div
                    className="timeline-group__children"
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: "auto", opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    transition={{ duration: 0.2 }}
                  >
                    {group.activities.map((act) => (
                      <div
                        key={act.id}
                        className={`timeline-child ${selectedActivity?.id === act.id ? "timeline-child--selected" : ""}`}
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedGroupKey(group.key);
                          setSelectedActivity(act);
                        }}
                      >
                        <div className="timeline-child__time">
                          {formatTime(act.timestamp ?? act.startedAt)}
                        </div>
                        <div className="timeline-child__title">{act.windowTitle}</div>
                        <div className="timeline-child__duration">
                          {formatSeconds(act.duration ?? act.durationSecs ?? 0)}
                        </div>
                      </div>
                    ))}
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
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
                <button className="icon-btn" onClick={() => { setSelectedActivity(null); setSelectedGroupKey(null); }}>
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
                  <span>{formatTime(selected.timestamp ?? selected.startedAt)} — {formatTime((selected.timestamp ?? 0) + (selected.duration ?? selected.durationSecs ?? 0))}</span>
                </div>
                <div className="meta-row">
                  <span className="meta-label">时长</span>
                  <span>{formatSeconds(selected.duration ?? selected.durationSecs ?? 0)}</span>
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

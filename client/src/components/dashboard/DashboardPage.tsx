import { useEffect } from "react";
import { motion } from "framer-motion";
import { Activity, Clock, Zap, Monitor } from "lucide-react";
import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { useInvoke } from "../../hooks/useInvoke";
import { CMD } from "../../utils/api";
import {
  formatDuration,
  categoryLabel,
  categoryColor,
} from "../../utils/format";
import type { DailyStats, HourlySummary } from "../../types";
import "./Dashboard.css";

// 动画变体 — 使用 any 绕过 framer-motion 严格泛型
const cardVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: i * 0.08, duration: 0.4, ease: "easeOut" },
  }),
} as any;

export default function DashboardPage() {
  const stats = useInvoke<DailyStats>(CMD.GET_TODAY_STATS);
  const hourly = useInvoke<HourlySummary[]>(CMD.GET_HOURLY_SUMMARIES);

  useEffect(() => {
    stats.execute();
    hourly.execute({ date: new Date().toISOString().slice(0, 10) });
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // 骨架屏
  if (stats.loading && !stats.data) {
    return <DashboardSkeleton />;
  }

  // 内联错误提示（不替换整个页面布局）
  const dataError = stats.error || hourly.error;

  const raw = stats.data;
  const hours = hourly.data ?? [];

  // 服务端返回 totalDuration(秒)，前端转分钟
  const totalMins = raw ? (raw.totalDuration != null ? raw.totalDuration / 60 : raw.totalMinutes ?? 0) : 0;
  const apps = raw?.appUsage?.map(a => ({ appName: a.appName, minutes: a.duration / 60, count: a.count })) ?? raw?.topApps ?? [];
  const cats = raw?.categoryUsage?.map(c => ({ category: c.category, minutes: c.duration / 60, percentage: 0 })) ?? raw?.categoryBreakdown ?? [];
  // 计算百分比
  const catTotal = cats.reduce((s, c) => s + c.minutes, 0) || 1;
  const catsWithPct = cats.map(c => ({ ...c, percentage: Math.round(c.minutes / catTotal * 100) }));
  const actCount = raw?.screenshotCount ?? raw?.totalActivities ?? apps.reduce((s, a) => s + (a.count ?? 0), 0);

  return (
    <div className="dashboard">
      {dataError && (
        <div className="dashboard__error-banner">
          <span>⚠ 数据加载异常：{dataError}</span>
          <button className="btn btn--sm" onClick={() => { stats.execute(); hourly.execute({ date: new Date().toISOString().slice(0, 10) }); }}>
            重试
          </button>
        </div>
      )}
      {/* Hero 数据卡片 */}
      <div className="dashboard__hero">
        <motion.div
          className="stat-card stat-card--accent"
          custom={0}
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          <div className="stat-card__icon">
            <Clock size={20} />
          </div>
          <div className="stat-card__label">今日专注时间</div>
          <div className="stat-card__value">
            {raw ? formatDuration(totalMins) : "--"}
          </div>
        </motion.div>

        <motion.div
          className="stat-card"
          custom={1}
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          <div className="stat-card__icon">
            <Activity size={20} />
          </div>
          <div className="stat-card__label">活动记录数</div>
          <div className="stat-card__value">
            {raw ? actCount : "--"}
          </div>
        </motion.div>

        <motion.div
          className="stat-card"
          custom={2}
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          <div className="stat-card__icon">
            <Zap size={20} />
          </div>
          <div className="stat-card__label">最常用应用</div>
          <div className="stat-card__value stat-card__value--sm">
            {apps[0]?.appName ?? "--"}
          </div>
        </motion.div>

        <motion.div
          className="stat-card"
          custom={3}
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          <div className="stat-card__icon">
            <Monitor size={20} />
          </div>
          <div className="stat-card__label">活跃时段</div>
          <div className="stat-card__value stat-card__value--sm">
            {raw?.activeHours ? `${raw.activeHours}小时` : "--"}
          </div>
        </motion.div>
      </div>

      {/* 中间行：分类饼图 + 热力图 */}
      <div className="dashboard__charts">
        <motion.div
          className="card dashboard__chart-card"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3, duration: 0.4 }}
        >
          <div className="card__title">分类占比</div>
          <CategoryDonut items={catsWithPct} />
        </motion.div>

        <motion.div
          className="card dashboard__chart-card"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4, duration: 0.4 }}
        >
          <div className="card__title">每小时活跃度</div>
          <HourlyHeatmap hours={hours} />
        </motion.div>
      </div>

      {/* 应用 TOP 列表 */}
      <motion.div
        className="card"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5, duration: 0.4 }}
      >
        <div className="card__title">应用排行</div>
        <div className="app-ranking">
          {apps.map((app, i) => (
            <div key={app.appName} className="app-ranking__item">
              <span className="app-ranking__rank">#{i + 1}</span>
              <span className="app-ranking__name">{app.appName}</span>
              <div className="app-ranking__bar-track">
                <motion.div
                  className="app-ranking__bar-fill"
                  initial={{ width: 0 }}
                  animate={{
                    width: `${totalMins ? (app.minutes / totalMins) * 100 : 0}%`,
                  }}
                  transition={{ delay: 0.6 + i * 0.05, duration: 0.6 }}
                />
              </div>
              <span className="app-ranking__time">
                {formatDuration(app.minutes)}
              </span>
            </div>
          ))}
          {apps.length === 0 && (
            <div className="app-ranking__empty">暂无数据</div>
          )}
        </div>
      </motion.div>
    </div>
  );
}

// ===== 子组件 =====

function CategoryDonut({
  items,
}: {
  items: { category: string; minutes: number; percentage: number }[];
}) {
  if (items.length === 0) {
    return <div className="chart-empty">暂无分类数据</div>;
  }

  const chartData = items.map((it) => ({
    name: categoryLabel(it.category),
    value: it.minutes,
    color: categoryColor(it.category),
  }));

  return (
    <div className="donut-wrapper">
      <ResponsiveContainer width="100%" height={220}>
        <PieChart>
          <Pie
            data={chartData}
            cx="50%"
            cy="50%"
            innerRadius={55}
            outerRadius={85}
            paddingAngle={3}
            dataKey="value"
            strokeWidth={0}
          >
            {chartData.map((entry) => (
              <Cell key={entry.name} fill={entry.color} />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{
              background: "var(--color-bg-card)",
              border: "1px solid var(--color-border)",
              borderRadius: "var(--radius-sm)",
              color: "var(--color-text-primary)",
              fontSize: "13px",
            }}
            formatter={(value: any) => [`${formatDuration(Number(value))}`, "时长"]}
          />
        </PieChart>
      </ResponsiveContainer>
      <div className="donut-legend">
        {chartData.map((it) => (
          <div key={it.name} className="donut-legend__item">
            <span
              className="donut-legend__dot"
              style={{ background: it.color }}
            />
            <span>{it.name}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function HourlyHeatmap({ hours }: { hours: HourlySummary[] }) {
  // 补齐 24 小时
  const grid = Array.from({ length: 24 }, (_, i) => {
    const h = hours.find((x) => x.hour === i);
    return { hour: i, minutes: (h?.totalDuration != null ? h.totalDuration / 60 : h?.totalMinutes) ?? 0 };
  });

  const maxMinutes = Math.max(...grid.map((g) => g.minutes), 1);

  return (
    <div className="heatmap">
      {grid.map((cell) => (
        <div
          key={cell.hour}
          className="heatmap__cell"
          style={{
            opacity: cell.minutes > 0 ? 0.2 + (cell.minutes / maxMinutes) * 0.8 : 0.06,
          }}
          title={`${cell.hour}:00 — ${formatDuration(cell.minutes)}`}
        >
          <span className="heatmap__hour">{cell.hour}</span>
        </div>
      ))}
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="dashboard">
      <div className="dashboard__hero">
        {[0, 1, 2, 3].map((i) => (
          <div key={i} className="stat-card skeleton" />
        ))}
      </div>
      <div className="dashboard__charts">
        <div className="card dashboard__chart-card skeleton skeleton--lg" />
        <div className="card dashboard__chart-card skeleton skeleton--lg" />
      </div>
    </div>
  );
}

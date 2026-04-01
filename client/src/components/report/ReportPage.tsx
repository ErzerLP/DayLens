import { useEffect, useState, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { FileText, RefreshCw, ChevronLeft, ChevronRight, Calendar, Sparkles, CheckCircle, AlertCircle } from "lucide-react";
import ReactMarkdown from "react-markdown";
import { useInvoke } from "../../hooks/useInvoke";
import { CMD } from "../../utils/api";
import { today } from "../../utils/format";
import { useLogStore } from "../../stores/logStore";
import type { DailyReport } from "../../types";
import "./Report.css";

type GenPhase = "idle" | "sending" | "generating" | "saving" | "done" | "error";

const PHASE_LABELS: Record<GenPhase, string> = {
  idle: "",
  sending: "请求已发送，正在连接 AI...",
  generating: "AI 正在分析工作数据，生成报告中...",
  saving: "报告生成完成，正在保存...",
  done: "报告已就绪 ✓",
  error: "生成失败",
};

const PHASE_ICONS: Record<GenPhase, React.ReactNode> = {
  idle: null,
  sending: <RefreshCw size={16} className="spin" />,
  generating: <Sparkles size={16} className="pulse" />,
  saving: <RefreshCw size={16} className="spin" />,
  done: <CheckCircle size={16} />,
  error: <AlertCircle size={16} />,
};

/** 格式化日期为中文友好格式 */
function formatDateCN(dateStr: string) {
  const d = new Date(dateStr + "T00:00:00");
  const weekdays = ["日", "一", "二", "三", "四", "五", "六"];
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日 周${weekdays[d.getDay()]}`;
}

/** 偏移日期 */
function shiftDate(dateStr: string, days: number): string {
  const d = new Date(dateStr + "T00:00:00");
  d.setDate(d.getDate() + days);
  return d.toISOString().slice(0, 10);
}

export default function ReportPage() {
  const [date, setDate] = useState(today());
  const report = useInvoke<DailyReport | null>(CMD.GET_REPORT);
  const generate = useInvoke<DailyReport>(CMD.GENERATE_REPORT);
  const [phase, setPhase] = useState<GenPhase>("idle");
  const [errorMsg, setErrorMsg] = useState("");
  const [elapsed, setElapsed] = useState(0);
  const timerRef = useRef<number>(undefined);

  useEffect(() => {
    report.execute({ date });
    setPhase("idle");
    setErrorMsg("");
  }, [date]); // eslint-disable-line react-hooks/exhaustive-deps

  // 计时器
  useEffect(() => {
    if (phase === "sending" || phase === "generating") {
      setElapsed(0);
      timerRef.current = window.setInterval(() => setElapsed((e) => e + 1), 1000);
      // 2 秒后自动切到 generating 阶段
      if (phase === "sending") {
        const t = window.setTimeout(() => setPhase("generating"), 2000);
        return () => { window.clearTimeout(t); window.clearInterval(timerRef.current); };
      }
    } else {
      window.clearInterval(timerRef.current);
    }
    return () => window.clearInterval(timerRef.current);
  }, [phase]);

  const handleGenerate = async () => {
    setPhase("sending");
    setErrorMsg("");
    useLogStore.getState().addLog("info", "ai", "开始生成日报: " + date);

    try {
      const result = await generate.execute({ date, force: false });
      if (result) {
        setPhase("saving");
        useLogStore.getState().addLog("success", "ai", "日报生成成功: " + date);
        // 短暂停留在 saving 阶段
        await new Promise((r) => setTimeout(r, 500));
        await report.execute({ date });
        setPhase("done");
        setTimeout(() => setPhase("idle"), 2000);
      } else {
        setPhase("error");
        setErrorMsg(generate.error || "生成失败，请检查 AI 配置和服务器连接");
        useLogStore.getState().addLog("error", "ai", "日报生成失败: " + (generate.error || "未知错误"));
      }
    } catch (e) {
      setPhase("error");
      setErrorMsg(String(e));
      useLogStore.getState().addLog("error", "ai", "日报生成异常: " + e);
    }
  };

  const content = report.data;
  const isToday = date === today();
  const isGenerating = phase === "sending" || phase === "generating" || phase === "saving";

  return (
    <div className="report-page">
      {/* 日期选择栏 */}
      <div className="report-page__toolbar">
        <div className="report-date-picker">
          <button
            className="report-date-picker__arrow"
            onClick={() => setDate(shiftDate(date, -1))}
            title="前一天"
          >
            <ChevronLeft size={18} />
          </button>

          <label className="report-date-picker__display">
            <Calendar size={14} className="report-date-picker__icon" />
            <span className="report-date-picker__text">{formatDateCN(date)}</span>
            {isToday && <span className="report-date-picker__today-badge">今天</span>}
            <input
              type="date"
              className="report-date-picker__hidden-input"
              value={date}
              max={today()}
              onChange={(e) => setDate(e.target.value)}
            />
          </label>

          <button
            className="report-date-picker__arrow"
            onClick={() => setDate(shiftDate(date, 1))}
            disabled={isToday}
            title="后一天"
          >
            <ChevronRight size={18} />
          </button>
        </div>

        <button
          className="settings-btn settings-btn--primary"
          onClick={handleGenerate}
          disabled={isGenerating}
        >
          {isGenerating ? (
            <RefreshCw size={14} className="spin" />
          ) : (
            <Sparkles size={14} />
          )}
          {isGenerating ? "生成中…" : "生成报告"}
        </button>
      </div>

      {/* 生成进度条 */}
      <AnimatePresence>
        {phase !== "idle" && (
          <motion.div
            className={`report-progress report-progress--${phase}`}
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.25 }}
          >
            <div className="report-progress__inner">
              <span className="report-progress__icon">{PHASE_ICONS[phase]}</span>
              <span className="report-progress__label">{PHASE_LABELS[phase]}</span>
              {(phase === "sending" || phase === "generating") && (
                <span className="report-progress__elapsed">{elapsed}s</span>
              )}
            </div>
            {phase === "error" && errorMsg && (
              <div className="report-progress__error">{errorMsg}</div>
            )}
            {(phase === "sending" || phase === "generating") && (
              <div className="report-progress__bar">
                <motion.div
                  className="report-progress__bar-fill"
                  initial={{ width: "0%" }}
                  animate={{ width: phase === "generating" ? "85%" : "20%" }}
                  transition={{ duration: phase === "generating" ? 30 : 2, ease: "linear" }}
                />
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>

      {/* 报告内容 */}
      <motion.div
        className="report-page__content card"
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        key={date}
      >
        {content ? (
          <div className="report-markdown">
            <ReactMarkdown>{content.content}</ReactMarkdown>
            <div className="report-page__generated-at">
              {content.usedAi && <span className="report-page__ai-badge">AI 生成</span>}
              生成时间：{content.generatedAt}
            </div>
          </div>
        ) : report.loading ? (
          <div className="report-page__empty">
            <RefreshCw size={24} className="spin" style={{ opacity: 0.4 }} />
            <p>加载中…</p>
          </div>
        ) : (
          <div className="report-page__empty">
            <FileText size={40} strokeWidth={1} />
            <p>该日暂无报告</p>
            <p className="report-page__hint">点击「生成报告」由 AI 自动分析当天工作情况</p>
          </div>
        )}
      </motion.div>
    </div>
  );
}

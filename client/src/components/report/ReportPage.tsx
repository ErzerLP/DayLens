import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { FileText, RefreshCw } from "lucide-react";
import ReactMarkdown from "react-markdown";
import { useInvoke } from "../../hooks/useInvoke";
import { CMD } from "../../utils/api";
import { today } from "../../utils/format";
import type { DailyReport } from "../../types";
import "./Report.css";

export default function ReportPage() {
  const [date, setDate] = useState(today());
  const report = useInvoke<DailyReport | null>(CMD.GET_REPORT);
  const generate = useInvoke<DailyReport>(CMD.GENERATE_REPORT);

  useEffect(() => {
    report.execute({ date });
  }, [date]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleGenerate = async () => {
    const result = await generate.execute({ date, force: false });
    if (result) {
      report.execute({ date });
    }
  };

  const content = report.data;

  return (
    <div className="report-page">
      <div className="report-page__toolbar">
        <input
          type="date"
          className="date-input"
          value={date}
          max={today()}
          onChange={(e) => setDate(e.target.value)}
        />
        <button
          className="btn btn--accent"
          onClick={handleGenerate}
          disabled={generate.loading}
        >
          <RefreshCw size={14} className={generate.loading ? "spin" : ""} />
          {generate.loading ? "生成中…" : "生成报告"}
        </button>
      </div>

      <motion.div
        className="report-page__content card"
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        {content ? (
          <div className="report-markdown">
            <ReactMarkdown>{content.content}</ReactMarkdown>
            <div className="report-page__generated-at">
              生成时间：{content.generatedAt}
            </div>
          </div>
        ) : report.loading ? (
          <div className="report-page__empty">加载中…</div>
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

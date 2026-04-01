// 应用分类管理页面

import { useState, useEffect, useRef, useCallback } from "react";
import ReactDOM from "react-dom";
import { invoke } from "@tauri-apps/api/core";
import { motion, AnimatePresence } from "framer-motion";
import { CMD } from "../../utils/api";
import { AppCategoryInfo, CATEGORIES, getCategoryInfo } from "../../types";
import { formatSeconds } from "../../utils/format";
import { useLogStore } from "../../stores/logStore";
import "./AppCategory.css";

type FilterMode = "all" | "other" | "custom";

interface PendingChange {
  appName: string;
  oldCategory: string;
  newCategory: string;
}

export default function AppCategoryPage() {
  const [apps, setApps] = useState<AppCategoryInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<FilterMode>("all");
  const [openDropdown, setOpenDropdown] = useState<string | null>(null);
  const [pendingChanges, setPendingChanges] = useState<PendingChange[]>([]);
  const [applying, setApplying] = useState(false);

  const fetchApps = useCallback(async () => {
    try {
      setLoading(true);
      const data = await invoke<AppCategoryInfo[]>(CMD.GET_APP_CATEGORIES);
      setApps(data);
      setError(null);
      setPendingChanges([]);
      useLogStore.getState().addLog("info", "data", `应用分类加载完成 — ${data.length} 个应用, ${data.filter(a => a.isCustomRule).length} 个自定义规则`);
    } catch (e) {
      setError(String(e));
      useLogStore.getState().addLog("error", "data", "应用分类加载失败: " + e);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchApps();
  }, [fetchApps]);

  // 选择新分类（仅标记，不立即提交）
  const handleSelectCategory = (appName: string, newCategory: string) => {
    const app = apps.find((a) => a.appName === appName);
    if (!app) return;

    // 如果选回了原分类，移除pending
    const originalCategory = app.category;
    setPendingChanges((prev) => {
      const filtered = prev.filter((c) => c.appName !== appName);
      if (newCategory !== originalCategory) {
        filtered.push({ appName, oldCategory: originalCategory, newCategory });
      }
      return filtered;
    });
    setOpenDropdown(null);
  };

  // 批量应用所有待定更改
  const handleApplyAll = async () => {
    if (pendingChanges.length === 0) return;
    setApplying(true);
    setError(null);
    try {
      for (const change of pendingChanges) {
        // 先设置规则
        await invoke(CMD.SET_CATEGORY_RULE, {
          appName: change.appName,
          category: change.newCategory,
        });
        // 再重分类历史数据
        await invoke(CMD.RECLASSIFY_APP, {
          appName: change.appName,
          newCategory: change.newCategory,
        });
        useLogStore.getState().addLog("success", "data", `分类已更新: ${change.appName}`, `${getCategoryInfo(change.oldCategory).label} → ${getCategoryInfo(change.newCategory).label}`);
      }
      // 重新拉取数据确认
      await fetchApps();
    } catch (e) {
      setError(String(e));
      useLogStore.getState().addLog("error", "data", "分类变更失败: " + e);
    } finally {
      setApplying(false);
    }
  };

  // 取消所有待定更改
  const handleCancelAll = () => {
    setPendingChanges([]);
  };

  // 获取某应用当前展示的分类（优先展示 pending）
  const getDisplayCategory = (appName: string): string => {
    const pending = pendingChanges.find((c) => c.appName === appName);
    return pending ? pending.newCategory : (apps.find((a) => a.appName === appName)?.category ?? "other");
  };

  // 过滤 & 搜索
  const filtered = apps
    .filter((a) => {
      const displayCat = getDisplayCategory(a.appName);
      if (filter === "other") return displayCat === "other";
      if (filter === "custom") return a.isCustomRule;
      return true;
    })
    .filter(
      (a) =>
        !search || a.appName.toLowerCase().includes(search.toLowerCase())
    )
    .sort((a, b) => b.totalDuration - a.totalDuration);

  const otherCount = apps.filter((a) => getDisplayCategory(a.appName) === "other").length;
  const customCount = apps.filter((a) => a.isCustomRule).length;
  const hasPending = pendingChanges.length > 0;

  return (
    <div className="app-category">
      <div className="app-category__header">
        <h1 className="app-category__title">应用分类管理</h1>
        <div className="app-category__toolbar">
          <input
            className="app-category__search"
            placeholder="搜索应用..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <div className="app-category__filter">
            {(
              [
                ["all", "全部"],
                ["other", `未分类 (${otherCount})`],
                ["custom", `自定义 (${customCount})`],
              ] as const
            ).map(([key, label]) => (
              <button
                key={key}
                className={`app-category__filter-btn ${
                  filter === key ? "app-category__filter-btn--active" : ""
                }`}
                onClick={() => setFilter(key)}
              >
                {label}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="app-category__stats">
        <span>
          共 <strong>{apps.length}</strong> 个应用
        </span>
        <span>
          未分类 <strong>{otherCount}</strong> 个
        </span>
        <span>
          自定义规则 <strong>{customCount}</strong> 个
        </span>
      </div>

      {/* 待定更改提示栏 */}
      {hasPending && (
        <div className="app-category__pending-bar">
          <span>
            📝 <strong>{pendingChanges.length}</strong> 项分类变更待确认
            {pendingChanges.map((c) => (
              <span key={c.appName} className="app-category__pending-item">
                {c.appName}: {getCategoryInfo(c.oldCategory).label} → {getCategoryInfo(c.newCategory).label}
              </span>
            ))}
          </span>
          <div className="app-category__pending-actions">
            <button
              className="btn btn--sm btn--ghost"
              onClick={handleCancelAll}
              disabled={applying}
            >
              取消
            </button>
            <button
              className="btn btn--sm btn--primary"
              onClick={handleApplyAll}
              disabled={applying}
            >
              {applying ? "应用中..." : `确认应用 (${pendingChanges.length})`}
            </button>
          </div>
        </div>
      )}

      {error && (
        <div className="dashboard__error-banner">
          <span>⚠ {error}</span>
          <button className="btn btn--sm" onClick={fetchApps}>
            重试
          </button>
        </div>
      )}

      {loading ? (
        <div className="app-category__empty">加载中...</div>
      ) : filtered.length === 0 ? (
        <div className="app-category__empty">
          {search ? "未找到匹配的应用" : "暂无应用数据"}
        </div>
      ) : (
        <table className="app-category__table">
          <thead>
            <tr>
              <th>应用名</th>
              <th>分类</th>
              <th>使用时长</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <AnimatePresence>
              {filtered.map((app) => {
                const displayCat = getDisplayCategory(app.appName);
                const isPending = pendingChanges.some((c) => c.appName === app.appName);
                return (
                  <motion.tr
                    key={app.appName}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className={isPending ? "app-category__row--pending" : ""}
                  >
                    <td>
                      <div className="app-category__app-name">
                        {app.appName}
                        {app.isCustomRule && (
                          <span className="app-category__custom-badge">
                            自定义
                          </span>
                        )}
                      </div>
                    </td>
                    <td>
                      <CategorySelector
                        appName={app.appName}
                        current={displayCat}
                        isOpen={openDropdown === app.appName}
                        onToggle={() =>
                          setOpenDropdown(
                            openDropdown === app.appName ? null : app.appName
                          )
                        }
                        onSelect={(cat) =>
                          handleSelectCategory(app.appName, cat)
                        }
                        disabled={applying}
                      />
                    </td>
                    <td className="app-category__duration">
                      {formatSeconds(app.totalDuration)}
                    </td>
                    <td>
                      {isPending ? (
                        <span className="app-category__pending-label">待确认</span>
                      ) : app.isCustomRule ? (
                        <span className="app-category__saved-label">✓ 已保存</span>
                      ) : (
                        <span className="app-category__auto-label">自动</span>
                      )}
                    </td>
                  </motion.tr>
                );
              })}
            </AnimatePresence>
          </tbody>
        </table>
      )}
    </div>
  );
}

// ===== 分类选择器组件 =====

interface CategorySelectorProps {
  appName: string;
  current: string;
  isOpen: boolean;
  onToggle: () => void;
  onSelect: (category: string) => void;
  disabled: boolean;
}

function CategorySelector({
  current,
  isOpen,
  onToggle,
  onSelect,
  disabled,
}: CategorySelectorProps) {
  const btnRef = useRef<HTMLButtonElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState({ top: 0, left: 0 });
  const info = getCategoryInfo(current);

  // 计算下拉框位置
  useEffect(() => {
    if (!isOpen || !btnRef.current) return;
    const rect = btnRef.current.getBoundingClientRect();
    setPos({
      top: rect.bottom + 4,
      left: rect.left,
    });
  }, [isOpen]);

  // 点击外部关闭
  useEffect(() => {
    if (!isOpen) return;
    const handler = (e: MouseEvent) => {
      const target = e.target as Node;
      if (
        btnRef.current && !btnRef.current.contains(target) &&
        dropdownRef.current && !dropdownRef.current.contains(target)
      ) {
        onToggle();
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [isOpen, onToggle]);

  return (
    <>
      <button
        ref={btnRef}
        className="category-badge"
        onClick={onToggle}
        disabled={disabled}
      >
        <span
          className="category-badge__dot"
          style={{ background: info.color }}
        />
        {info.label}
        <span className="category-badge__chevron">▾</span>
      </button>

      {isOpen &&
        ReactDOM.createPortal(
          <div
            ref={dropdownRef}
            className="category-dropdown"
            style={{
              position: "fixed",
              top: pos.top,
              left: pos.left,
            }}
          >
            {CATEGORIES.map((cat) => (
              <button
                key={cat.key}
                className={`category-dropdown__item ${
                  cat.key === current ? "category-dropdown__item--active" : ""
                }`}
                onClick={() => onSelect(cat.key)}
              >
                <span
                  className="category-badge__dot"
                  style={{ background: cat.color }}
                />
                {cat.label}
              </button>
            ))}
          </div>,
          document.body
        )}
    </>
  );
}

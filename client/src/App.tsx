import { useEffect } from "react";
import { BrowserRouter, Routes, Route, NavLink, useLocation, useNavigate } from "react-router-dom";
import { AnimatePresence, motion } from "framer-motion";
import { ToastProvider } from "./components/common/Toast";
import {
  LayoutDashboard,
  Clock,
  FileText,
  Search,
  Settings,
  RefreshCw,
  AlertTriangle,
  ScrollText,
  Tags,
} from "lucide-react";
import { useSystemStore } from "./stores/systemStore";
import DashboardPage from "./components/dashboard/DashboardPage";
import TimelinePage from "./components/timeline/TimelinePage";
import ReportPage from "./components/report/ReportPage";
import SearchPage from "./components/search/SearchPage";
import SettingsPage from "./components/settings/SettingsPage";
import LogPage from "./components/logs/LogPage";
import AppCategoryPage from "./components/categories/AppCategoryPage";
import "./index.css";

// ===== Sidebar =====

interface NavItemProps {
  to: string;
  icon: React.ReactNode;
  label: string;
}

function SidebarNavItem({ to, icon, label }: NavItemProps) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `sidebar__nav-item ${isActive ? "sidebar__nav-item--active" : ""}`
      }
    >
      {icon}
      {label}
    </NavLink>
  );
}

function Sidebar() {
  const { syncQueueSize, fetchSyncQueue, fetchConfig, checkConnection, isServerConnected } = useSystemStore();

  useEffect(() => {
    fetchConfig();
    fetchSyncQueue();
    checkConnection();
    const timer = setInterval(() => {
      fetchSyncQueue();
      checkConnection();
    }, 30_000);
    return () => clearInterval(timer);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <aside className="app-layout__sidebar">
      <div className="sidebar__logo">
        <div className="sidebar__logo-icon">D</div>
        DayLens
      </div>

      <div className="sidebar__section-title">概览</div>
      <SidebarNavItem to="/" icon={<LayoutDashboard size={18} />} label="仪表盘" />
      <SidebarNavItem to="/timeline" icon={<Clock size={18} />} label="时间线" />

      <div className="sidebar__section-title">分析</div>
      <SidebarNavItem to="/report" icon={<FileText size={18} />} label="日报" />
      <SidebarNavItem to="/search" icon={<Search size={18} />} label="搜索 / AI" />

      <div className="sidebar__section-title">系统</div>
      <SidebarNavItem to="/settings" icon={<Settings size={18} />} label="设置" />
      <SidebarNavItem to="/categories" icon={<Tags size={18} />} label="分类管理" />
      <SidebarNavItem to="/logs" icon={<ScrollText size={18} />} label="日志" />

      {/* 底部状态 */}
      <div className="sidebar__spacer" />
      <div className="sidebar__status">
        {!isServerConnected && (
          <div className="sidebar__connection-badge sidebar__connection-badge--warn">
            <AlertTriangle size={12} />
            未连接
          </div>
        )}
        {syncQueueSize > 0 && (
          <div className="sidebar__sync-badge">
            <RefreshCw size={12} className="spin" />
            {syncQueueSize} 待同步
          </div>
        )}
      </div>
    </aside>
  );
}

// ===== Header =====

const PAGE_TITLES: Record<string, string> = {
  "/": "仪表盘",
  "/timeline": "时间线",
  "/report": "日报",
  "/search": "搜索 / AI",
  "/settings": "设置",
  "/logs": "日志",
};

function Header() {
  const location = useLocation();
  const navigate = useNavigate();
  const title = PAGE_TITLES[location.pathname] ?? "DayLens";
  const isServerConnected = useSystemStore((s) => s.isServerConnected);

  return (
    <>
      <header className="app-layout__header">
        <h1 className="header__title">{title}</h1>
        <div className="header__spacer" />
        <div className="header__status">
          {isServerConnected ? (
            <>
              <span className="header__status-dot header__status-dot--connected animate-pulse" />
              已连接
            </>
          ) : (
            <>
              <span className="header__status-dot header__status-dot--disconnected" />
              未连接
            </>
          )}
        </div>
      </header>
      {!isServerConnected && (
        <div className="connection-banner">
          <AlertTriangle size={14} />
          <span>服务端未连接，数据无法同步。</span>
          <button className="connection-banner__btn" onClick={() => navigate("/settings")}>
            前往设置
          </button>
        </div>
      )}
    </>
  );
}

// ===== 路由动画包装 =====

function AnimatedRoutes() {
  const location = useLocation();
  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={location.pathname}
        initial={{ opacity: 0, y: 6 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -6 }}
        transition={{ duration: 0.2 }}
        style={{ height: "100%" }}
      >
        <Routes location={location}>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/timeline" element={<TimelinePage />} />
          <Route path="/report" element={<ReportPage />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/logs" element={<LogPage />} />
          <Route path="/categories" element={<AppCategoryPage />} />
        </Routes>
      </motion.div>
    </AnimatePresence>
  );
}

// ===== 根组件 =====

export default function App() {
  return (
    <ToastProvider>
      <BrowserRouter>
        <div className="app-layout">
          <Sidebar />
          <div className="app-layout__main">
            <Header />
            <main className="app-layout__content">
              <AnimatedRoutes />
            </main>
          </div>
        </div>
      </BrowserRouter>
    </ToastProvider>
  );
}

import { useEffect } from "react";
import { BrowserRouter, Routes, Route, NavLink, useLocation } from "react-router-dom";
import { AnimatePresence, motion } from "framer-motion";
import { ToastProvider } from "./components/common/Toast";
import {
  LayoutDashboard,
  Clock,
  FileText,
  Search,
  Settings,
  RefreshCw,
} from "lucide-react";
import { useSystemStore } from "./stores/systemStore";
import DashboardPage from "./components/dashboard/DashboardPage";
import TimelinePage from "./components/timeline/TimelinePage";
import ReportPage from "./components/report/ReportPage";
import SearchPage from "./components/search/SearchPage";
import SettingsPage from "./components/settings/SettingsPage";
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
  const { syncQueueSize, fetchSyncQueue, fetchConfig } = useSystemStore();

  useEffect(() => {
    fetchConfig();
    fetchSyncQueue();
    const timer = setInterval(fetchSyncQueue, 30_000);
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

      {/* 底部状态 */}
      <div className="sidebar__spacer" />
      <div className="sidebar__status">
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
};

function Header() {
  const location = useLocation();
  const title = PAGE_TITLES[location.pathname] ?? "DayLens";
  const config = useSystemStore((s) => s.config);

  return (
    <header className="app-layout__header">
      <h1 className="header__title">{title}</h1>
      <div className="header__spacer" />
      <div className="header__status">
        {config ? (
          <>
            <span className="header__status-dot header__status-dot--connected animate-pulse" />
            已配置
          </>
        ) : (
          <>
            <span className="header__status-dot header__status-dot--disconnected" />
            未连接
          </>
        )}
      </div>
    </header>
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

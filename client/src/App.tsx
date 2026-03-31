import { useEffect, lazy, Suspense } from "react";
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
import { useThemeStore, syncThemeToDom, setupSystemThemeListener } from "./stores/themeStore";
import "./index.css";

// ===== 路由级懒加载 =====

const DashboardPage = lazy(() => import("./components/dashboard/DashboardPage"));
const TimelinePage = lazy(() => import("./components/timeline/TimelinePage"));
const ReportPage = lazy(() => import("./components/report/ReportPage"));
const SearchPage = lazy(() => import("./components/search/SearchPage"));
const SettingsPage = lazy(() => import("./components/settings/SettingsPage"));
const LogPage = lazy(() => import("./components/logs/LogPage"));
const AppCategoryPage = lazy(() => import("./components/categories/AppCategoryPage"));

// ===== 页面加载占位 =====

function PageSkeleton() {
  return (
    <div className="page-placeholder animate-fade-in">
      <div className="page-placeholder__text">加载中…</div>
    </div>
  );
}

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
  // 细粒度订阅 — 只订阅需要渲染的数据字段
  const syncQueueSize = useSystemStore((s) => s.syncQueueSize);
  const isServerConnected = useSystemStore((s) => s.isServerConnected);

  useEffect(() => {
    // 方法引用是稳定的，直接从 store 获取
    const { fetchConfig, fetchSyncQueue, checkConnection } = useSystemStore.getState();
    fetchConfig();
    fetchSyncQueue();
    checkConnection();
    const timer = setInterval(() => {
      fetchSyncQueue();
      checkConnection();
    }, 30_000);
    return () => clearInterval(timer);
  }, []);

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

// Material Design 3 动效缓动函数
const MD3_EASING: [number, number, number, number] = [0.2, 0, 0, 1]; // Standard / Decelerate

function AnimatedRoutes() {
  const location = useLocation();

  // MD3 Fade Through (淡入穿透)
  const fadeThroughVariants = {
    initial: { opacity: 0, scale: 0.97, y: 4 },
    animate: { opacity: 1, scale: 1, y: 0 },
    exit: { opacity: 0, scale: 1.02, y: -4 },
  };

  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={location.pathname}
        variants={fadeThroughVariants}
        initial="initial"
        animate="animate"
        exit="exit"
        transition={{ duration: 0.25, ease: MD3_EASING }}
        style={{ height: "100%", width: "100%", display: "flex", flexDirection: "column" }}
      >
        <Suspense fallback={<PageSkeleton />}>
          <Routes location={location}>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/timeline" element={<TimelinePage />} />
            <Route path="/report" element={<ReportPage />} />
            <Route path="/search" element={<SearchPage />} />
            <Route path="/settings" element={<SettingsPage />} />
            <Route path="/logs" element={<LogPage />} />
            <Route path="/categories" element={<AppCategoryPage />} />
          </Routes>
        </Suspense>
      </motion.div>
    </AnimatePresence>
  );
}

// ===== 根组件 =====

export default function App() {
  const theme = useThemeStore((s) => s.theme);

  useEffect(() => {
    syncThemeToDom(theme);
    const cleanup = setupSystemThemeListener();
    return cleanup;
  }, [theme]);

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

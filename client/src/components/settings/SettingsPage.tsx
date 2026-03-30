import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { Server, Camera, HardDrive, Save, RefreshCw } from "lucide-react";
import { useSystemStore } from "../../stores/systemStore";
import { formatBytes } from "../../utils/format";
import "./Settings.css";

type Section = "server" | "capture" | "storage";

export default function SettingsPage() {
  const [section, setSection] = useState<Section>("server");

  return (
    <div className="settings-page">
      <div className="settings-page__nav">
        <button
          className={`settings-nav-item ${section === "server" ? "settings-nav-item--active" : ""}`}
          onClick={() => setSection("server")}
        >
          <Server size={16} /> 服务器
        </button>
        <button
          className={`settings-nav-item ${section === "capture" ? "settings-nav-item--active" : ""}`}
          onClick={() => setSection("capture")}
        >
          <Camera size={16} /> 采集设置
        </button>
        <button
          className={`settings-nav-item ${section === "storage" ? "settings-nav-item--active" : ""}`}
          onClick={() => setSection("storage")}
        >
          <HardDrive size={16} /> 存储管理
        </button>
      </div>

      <motion.div
        className="settings-page__content"
        key={section}
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.2 }}
      >
        {section === "server" && <ServerSection />}
        {section === "capture" && <CaptureSection />}
        {section === "storage" && <StorageSection />}
      </motion.div>
    </div>
  );
}

// ===== 服务器设置 =====

function ServerSection() {
  const { config, fetchConfig, updateServerUrl, updateServerToken } = useSystemStore();
  const [url, setUrl] = useState("");
  const [token, setToken] = useState("");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    fetchConfig();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (config) {
      setUrl(config.server.url);
      setToken(config.server.token);
    }
  }, [config]);

  const handleSave = async () => {
    try {
      await updateServerUrl(url);
      await updateServerToken(token);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (e) {
      console.error("保存失败:", e);
    }
  };

  return (
    <div className="settings-section">
      <h3 className="settings-section__title">服务器连接</h3>

      <div className="form-group">
        <label className="form-label">服务器地址</label>
        <input
          className="form-input"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="http://localhost:8080"
        />
      </div>

      <div className="form-group">
        <label className="form-label">认证 Token</label>
        <input
          className="form-input"
          type="password"
          value={token}
          onChange={(e) => setToken(e.target.value)}
          placeholder="Bearer token"
        />
      </div>

      <button className="btn btn--accent" onClick={handleSave}>
        <Save size={14} />
        {saved ? "已保存 ✓" : "保存"}
      </button>
    </div>
  );
}

// ===== 采集设置 =====

function CaptureSection() {
  const { config, fetchConfig, updateCaptureInterval } = useSystemStore();
  const [interval, setInterval_] = useState(30);

  useEffect(() => {
    fetchConfig();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (config) setInterval_(config.capture.screenshotIntervalSecs);
  }, [config]);

  const handleSave = async () => {
    try {
      await updateCaptureInterval(interval);
    } catch (e) {
      console.error("保存失败:", e);
    }
  };

  return (
    <div className="settings-section">
      <h3 className="settings-section__title">采集参数</h3>

      <div className="form-group">
        <label className="form-label">截屏间隔（秒）</label>
        <input
          className="form-input form-input--narrow"
          type="number"
          min={5}
          max={300}
          value={interval}
          onChange={(e) => setInterval_(Number(e.target.value))}
        />
        <span className="form-hint">建议 15–60 秒，间隔越短数据越精细但占用更多存储</span>
      </div>

      <button className="btn btn--accent" onClick={handleSave}>
        <Save size={14} /> 保存
      </button>
    </div>
  );
}

// ===== 存储管理 =====

function StorageSection() {
  const { storageStats, fetchStorageStats } = useSystemStore();

  useEffect(() => {
    fetchStorageStats();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="settings-section">
      <h3 className="settings-section__title">存储统计</h3>

      <div className="storage-grid">
        <div className="storage-stat">
          <div className="storage-stat__label">总占用</div>
          <div className="storage-stat__value">
            {storageStats ? formatBytes(storageStats.totalBytes) : "--"}
          </div>
        </div>
        <div className="storage-stat">
          <div className="storage-stat__label">截图占用</div>
          <div className="storage-stat__value">
            {storageStats ? formatBytes(storageStats.screenshotBytes) : "--"}
          </div>
        </div>
        <div className="storage-stat">
          <div className="storage-stat__label">截图数量</div>
          <div className="storage-stat__value">
            {storageStats?.screenshotCount ?? "--"}
          </div>
        </div>
      </div>

      <button className="btn" onClick={fetchStorageStats}>
        <RefreshCw size={14} /> 刷新
      </button>
    </div>
  );
}

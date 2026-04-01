import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { Server, Camera, HardDrive, Save, RefreshCw, Wifi, Brain, Zap, ChevronDown, Power, Monitor, Moon, Sun } from "lucide-react";
import { invoke } from "@tauri-apps/api/core";
import { useSystemStore } from "../../stores/systemStore";
import { useThemeStore } from "../../stores/themeStore";
import { useLogStore } from "../../stores/logStore";
import "./Settings.css";

type Section = "server" | "capture" | "storage" | "ai" | "appearance";

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
          className={`settings-nav-item ${section === "ai" ? "settings-nav-item--active" : ""}`}
          onClick={() => setSection("ai")}
        >
          <Brain size={16} /> AI 模型
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
        <button
          className={`settings-nav-item ${section === "appearance" ? "settings-nav-item--active" : ""}`}
          onClick={() => setSection("appearance")}
        >
          <Monitor size={16} /> 外观
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
        {section === "ai" && <AISection />}
        {section === "capture" && <CaptureSection />}
        {section === "storage" && <StorageSection />}
        {section === "appearance" && <AppearanceSection />}
      </motion.div>
    </div>
  );
}

// ===== 服务器设置 =====

// 常用时区列表
const TIMEZONES = [
  { value: "Asia/Shanghai", label: "中国标准时间 (UTC+8)" },
  { value: "Asia/Tokyo", label: "日本标准时间 (UTC+9)" },
  { value: "Asia/Seoul", label: "韩国标准时间 (UTC+9)" },
  { value: "Asia/Singapore", label: "新加坡标准时间 (UTC+8)" },
  { value: "Asia/Hong_Kong", label: "香港标准时间 (UTC+8)" },
  { value: "Asia/Taipei", label: "台北标准时间 (UTC+8)" },
  { value: "Asia/Kolkata", label: "印度标准时间 (UTC+5:30)" },
  { value: "Europe/London", label: "英国时间 (UTC+0/+1)" },
  { value: "Europe/Berlin", label: "中欧时间 (UTC+1/+2)" },
  { value: "Europe/Moscow", label: "莫斯科时间 (UTC+3)" },
  { value: "America/New_York", label: "美国东部时间 (UTC-5/-4)" },
  { value: "America/Chicago", label: "美国中部时间 (UTC-6/-5)" },
  { value: "America/Los_Angeles", label: "美国太平洋时间 (UTC-8/-7)" },
  { value: "Pacific/Auckland", label: "新西兰标准时间 (UTC+12/+13)" },
  { value: "Australia/Sydney", label: "澳大利亚东部时间 (UTC+10/+11)" },
  { value: "UTC", label: "协调世界时 (UTC)" },
];

function ServerSection() {
  const { config, fetchConfig, updateServerUrl, updateServerToken, testConnection, isServerConnected, connectionChecking } = useSystemStore();
  const [url, setUrl] = useState("");
  const [token, setToken] = useState("");
  const [timezone, setTimezone] = useState("Asia/Shanghai");
  const [saved, setSaved] = useState(false);
  const [testResult, setTestResult] = useState<"idle" | "success" | "fail">("idle");
  const [tzSynced, setTzSynced] = useState(false);

  useEffect(() => {
    fetchConfig();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (config) {
      setUrl(config.server.url);
      setToken(config.server.token);
    }
  }, [config]);

  // 从服务器加载时区
  useEffect(() => {
    if (!config?.server.url || !config?.server.token) return;
    serverFetch<{ timezone: string }>("/api/v1/config/timezone")
      .then((data) => {
        if (data.timezone) setTimezone(data.timezone);
      })
      .catch(() => {});
  }, [config?.server.url, config?.server.token]);

  const handleSave = async () => {
    try {
      await updateServerUrl(url);
      await updateServerToken(token);
      setSaved(true);
      setTestResult("idle");
      setTimeout(() => setSaved(false), 2000);
      useLogStore.getState().addLog("success", "config", "服务器配置已保存", `URL: ${url}`);
    } catch (e) {
      console.error("保存失败:", e);
      useLogStore.getState().addLog("error", "config", "服务器配置保存失败: " + e);
    }
  };

  const handleTest = async () => {
    try {
      await updateServerUrl(url);
      await updateServerToken(token);
    } catch {}
    const ok = await testConnection();
    setTestResult(ok ? "success" : "fail");
    setTimeout(() => setTestResult("idle"), 5000);
  };

  const handleSaveTimezone = async (tz: string) => {
    setTimezone(tz);
    try {
      await serverFetch("/api/v1/config/timezone", {
        method: "PUT",
        body: JSON.stringify({ timezone: tz }),
      });
      setTzSynced(true);
      setTimeout(() => setTzSynced(false), 2000);
      useLogStore.getState().addLog("success", "config", "时区已同步: " + tz);
    } catch (e) {
      console.error("时区同步失败:", e);
      useLogStore.getState().addLog("error", "config", "时区同步失败: " + e);
    }
  };

  return (
    <div className="settings-section">
      <h3 className="settings-section__title">服务器连接</h3>

      <div className={`connection-status ${isServerConnected ? "connection-status--ok" : "connection-status--fail"}`}>
        <span className={`connection-status__dot ${isServerConnected ? "connection-status__dot--ok" : "connection-status__dot--fail"}`} />
        {isServerConnected ? "已连接到服务器" : "未连接到服务器"}
      </div>

      <div className="form-group">
        <label className="form-label">服务器地址</label>
        <input className="form-input" value={url} onChange={(e) => setUrl(e.target.value)} placeholder="http://your-server:8080" />
      </div>

      <div className="form-group">
        <label className="form-label">认证 Token</label>
        <input className="form-input" type="password" value={token} onChange={(e) => setToken(e.target.value)} placeholder="Bearer token" />
      </div>

      <div className="settings-section__actions">
        <button className="settings-btn settings-btn--primary" onClick={handleSave}>
          <Save size={14} />
          {saved ? "已保存 ✓" : "保存"}
        </button>
        <button className="settings-btn settings-btn--outline" onClick={handleTest} disabled={connectionChecking || !url}>
          {connectionChecking ? <RefreshCw size={14} className="spin" /> : <Wifi size={14} />}
          {connectionChecking ? "测试中..." : "测试连接"}
        </button>
        {testResult === "success" && <span className="test-result test-result--ok">✓ 连接成功</span>}
        {testResult === "fail" && <span className="test-result test-result--fail">✗ 连接失败</span>}
      </div>

      <h3 className="settings-section__title" style={{ marginTop: "var(--space-lg)" }}>时区设置</h3>
      <p className="settings-section__desc">选择所在时区，服务端将按此时区统计每小时活跃度和工作时段。</p>

      <div className="form-group">
        <label className="form-label">时区</label>
        <select
          className="form-input form-select"
          value={timezone}
          onChange={(e) => handleSaveTimezone(e.target.value)}
        >
          {TIMEZONES.map((tz) => (
            <option key={tz.value} value={tz.value}>
              {tz.label}
            </option>
          ))}
        </select>
      </div>
      {tzSynced && <span className="test-result test-result--ok">✓ 时区已同步到服务器</span>}
    </div>
  );
}

// ===== AI 模型设置 =====

interface AIProvider {
  id: string;
  name: string;
  defaultEndpoint: string;
  defaultModel: string;
  requiresApiKey: boolean;
}

interface AIConfig {
  provider: string;
  endpoint: string;
  model: string;
  apiKey: string;
  customPrompt: string;
}

async function serverFetch<T>(path: string, opts?: RequestInit): Promise<T> {
  const store = useSystemStore.getState();
  const url = store.config?.server.url;
  const token = store.config?.server.token;
  if (!url) throw new Error("未配置服务器地址");
  const resp = await fetch(`${url}${path}`, {
    ...opts,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...(opts?.headers ?? {}),
    },
  });
  if (!resp.ok) {
    const text = await resp.text();
    throw new Error(`${resp.status}: ${text}`);
  }
  const json = await resp.json();
  if (json.code !== 0) throw new Error(json.message || "服务端错误");
  return json.data as T;
}

/** 从 OpenAI 兼容 /models 接口获取模型列表 */
async function fetchModels(endpoint: string, apiKey: string): Promise<string[]> {
  try {
    const base = endpoint.replace(/\/+$/, "");
    const resp = await fetch(`${base}/models`, {
      headers: apiKey ? { Authorization: `Bearer ${apiKey}` } : {},
    });
    if (!resp.ok) return [];
    const json = await resp.json();
    // OpenAI 格式: { data: [{ id: "model-name" }] }
    if (json.data && Array.isArray(json.data)) {
      return json.data.map((m: { id: string }) => m.id).sort();
    }
    return [];
  } catch {
    return [];
  }
}

function AISection() {
  const { isServerConnected } = useSystemStore();
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [form, setForm] = useState<AIConfig>({
    provider: "gemini",
    endpoint: "",
    model: "",
    apiKey: "",
    customPrompt: "",
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; latencyMs: number; reply?: string; error?: string } | null>(null);
  const [error, setError] = useState<string | null>(null);

  // 模型列表
  const [modelList, setModelList] = useState<string[]>([]);
  const [modelDropdown, setModelDropdown] = useState(false);
  const [fetchingModels, setFetchingModels] = useState(false);

  useEffect(() => {
    if (!isServerConnected) return;
    (async () => {
      try {
        setLoading(true);
        const [provResp, cfgResp] = await Promise.all([
          serverFetch<{ items: AIProvider[] }>("/api/v1/config/ai/providers"),
          serverFetch<AIConfig>("/api/v1/config/ai"),
        ]);
        setProviders(provResp.items);
        setForm(cfgResp);
        setError(null);
      } catch (e) {
        setError(String(e));
      } finally {
        setLoading(false);
      }
    })();
  }, [isServerConnected]);

  const handleProviderChange = (id: string) => {
    const p = providers.find((pr) => pr.id === id);
    setForm((f) => ({
      ...f,
      provider: id,
      endpoint: p?.defaultEndpoint ?? "",
      model: p?.defaultModel ?? "",
    }));
    setTestResult(null);
    setModelList([]);
  };

  const handleFetchModels = async () => {
    if (!form.endpoint) return;
    setFetchingModels(true);
    const models = await fetchModels(form.endpoint, form.apiKey);
    setModelList(models);
    setFetchingModels(false);
    if (models.length > 0) setModelDropdown(true);
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    try {
      await serverFetch("/api/v1/config/ai", {
        method: "PUT",
        body: JSON.stringify(form),
      });
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
      useLogStore.getState().addLog("success", "ai", `AI 配置已保存: ${form.provider} / ${form.model}`);
    } catch (e) {
      setError(String(e));
      useLogStore.getState().addLog("error", "ai", "AI 配置保存失败: " + e);
    } finally {
      setSaving(false);
    }
  };

  const handleTest = async () => {
    setTesting(true);
    setTestResult(null);
    setError(null);
    useLogStore.getState().addLog("info", "ai", `正在测试 AI 连接: ${form.provider} / ${form.model}`);
    try {
      const result = await serverFetch<{ success: boolean; latencyMs: number; reply?: string; error?: string }>(
        "/api/v1/config/ai/test",
        {
          method: "POST",
          body: JSON.stringify({
            provider: form.provider,
            endpoint: form.endpoint,
            model: form.model,
            apiKey: form.apiKey,
          }),
        }
      );
      setTestResult(result);
      if (result.success) {
        useLogStore.getState().addLog("success", "ai", `AI 连接成功 (${result.latencyMs}ms)`, result.reply);
      } else {
        useLogStore.getState().addLog("error", "ai", "AI 连接失败", result.error);
      }
    } catch (e) {
      setTestResult({ success: false, latencyMs: 0, error: String(e) });
      useLogStore.getState().addLog("error", "ai", "AI 连接测试异常: " + e);
    } finally {
      setTesting(false);
    }
  };

  if (!isServerConnected) {
    return (
      <div className="settings-section">
        <h3 className="settings-section__title">AI 模型配置</h3>
        <div className="ai-not-connected">
          <Brain size={32} />
          <p>请先在「服务器」页面连接到服务端后再配置 AI</p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="settings-section">
        <h3 className="settings-section__title">AI 模型配置</h3>
        <div className="ai-not-connected">加载中...</div>
      </div>
    );
  }

  const currentProvider = providers.find((p) => p.id === form.provider);

  return (
    <div className="settings-section">
      <h3 className="settings-section__title">AI 模型配置</h3>
      <p className="settings-section__desc">配置服务端使用的大模型 API，用于日报生成、智能问答等功能</p>

      {error && (
        <div className="dashboard__error-banner" style={{ marginBottom: 12 }}>
          <span>⚠ {error}</span>
        </div>
      )}

      {/* Provider 选择 */}
      <div className="form-group">
        <label className="form-label">AI 提供商</label>
        <div className="ai-provider-grid">
          {providers.map((p) => (
            <button
              key={p.id}
              className={`ai-provider-card ${form.provider === p.id ? "ai-provider-card--active" : ""}`}
              onClick={() => handleProviderChange(p.id)}
            >
              <span className="ai-provider-card__name">{p.name}</span>
              {!p.requiresApiKey && <span className="ai-provider-card__badge">免费本地</span>}
            </button>
          ))}
        </div>
      </div>

      {/* Endpoint */}
      {currentProvider && currentProvider.defaultEndpoint && (
        <div className="form-group">
          <label className="form-label">API 端点</label>
          <input
            className="form-input"
            value={form.endpoint}
            onChange={(e) => setForm((f) => ({ ...f, endpoint: e.target.value }))}
            placeholder={currentProvider.defaultEndpoint}
          />
          <span className="form-hint">留空则使用默认端点</span>
        </div>
      )}

      {/* API Key */}
      {currentProvider?.requiresApiKey && (
        <div className="form-group">
          <label className="form-label">API Key</label>
          <input
            className="form-input"
            type="password"
            value={form.apiKey}
            onChange={(e) => setForm((f) => ({ ...f, apiKey: e.target.value }))}
            placeholder="sk-..."
          />
        </div>
      )}

      {/* Model — 带下拉选择 + 自定义输入 */}
      <div className="form-group">
        <label className="form-label">模型名称</label>
        <div className="model-selector">
          <input
            className="form-input model-selector__input"
            value={form.model}
            onChange={(e) => setForm((f) => ({ ...f, model: e.target.value }))}
            placeholder={currentProvider?.defaultModel ?? "输入模型名称"}
          />
          {form.endpoint && (
            <button
              className="model-selector__fetch-btn"
              onClick={handleFetchModels}
              disabled={fetchingModels}
              title="从 API 获取可用模型列表"
            >
              {fetchingModels ? <RefreshCw size={14} className="spin" /> : <ChevronDown size={14} />}
            </button>
          )}
        </div>
        {modelList.length > 0 && modelDropdown && (
          <div className="model-dropdown">
            {modelList.map((m) => (
              <button
                key={m}
                className={`model-dropdown__item ${form.model === m ? "model-dropdown__item--active" : ""}`}
                onClick={() => {
                  setForm((f) => ({ ...f, model: m }));
                  setModelDropdown(false);
                }}
              >
                {m}
              </button>
            ))}
            <button className="model-dropdown__close" onClick={() => setModelDropdown(false)}>
              关闭列表
            </button>
          </div>
        )}
        {modelList.length === 0 && !fetchingModels && form.endpoint && (
          <span className="form-hint">可直接输入模型名，或点击右侧按钮从 API 获取列表</span>
        )}
      </div>

      {/* 测试结果 */}
      {testResult && (
        <div className={`ai-test-result ${testResult.success ? "ai-test-result--ok" : "ai-test-result--fail"}`}>
          <div className="ai-test-result__header">
            {testResult.success ? "✓ 连接成功" : "✗ 连接失败"}
            <span className="ai-test-result__latency">{testResult.latencyMs}ms</span>
          </div>
          {testResult.reply && <div className="ai-test-result__reply">回复: {testResult.reply}</div>}
          {testResult.error && <div className="ai-test-result__error">{testResult.error}</div>}
        </div>
      )}

      <div className="settings-section__actions">
        <button className="settings-btn settings-btn--primary" onClick={handleSave} disabled={saving}>
          <Save size={14} />
          {saved ? "已保存 ✓" : saving ? "保存中..." : "保存配置"}
        </button>
        <button className="settings-btn settings-btn--outline" onClick={handleTest} disabled={testing}>
          {testing ? <RefreshCw size={14} className="spin" /> : <Zap size={14} />}
          {testing ? "测试中..." : "测试连接"}
        </button>
      </div>
    </div>
  );
}

// ===== 采集设置 =====

function CaptureSection() {
  const { config, fetchConfig, updateCaptureInterval } = useSystemStore();
  const [interval, setInterval_] = useState(30);
  const [autostart, setAutostart] = useState(false);
  const [autostartLoading, setAutostartLoading] = useState(true);

  useEffect(() => {
    fetchConfig();
    // 查询开机自启状态
    invoke<boolean>("is_autostart_enabled")
      .then(setAutostart)
      .catch(() => setAutostart(false))
      .finally(() => setAutostartLoading(false));
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

  const handleToggleAutostart = async () => {
    try {
      if (autostart) {
        await invoke("disable_autostart");
        setAutostart(false);
      } else {
        await invoke("enable_autostart");
        setAutostart(true);
      }
    } catch (e) {
      console.error("开机自启设置失败:", e);
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

      <h3 className="settings-section__title" style={{ marginTop: "var(--space-lg)" }}>系统设置</h3>

      <div className="form-group autostart-row">
        <div className="autostart-row__info">
          <Power size={16} />
          <div>
            <span className="form-label" style={{ marginBottom: 0 }}>开机自启动</span>
            <span className="form-hint">系统启动时自动运行 DayLens</span>
          </div>
        </div>
        <button
          className={`toggle-switch ${autostart ? "toggle-switch--on" : ""}`}
          onClick={handleToggleAutostart}
          disabled={autostartLoading}
          aria-label="开机自启开关"
        >
          <span className="toggle-switch__knob" />
        </button>
      </div>
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
          <div className="storage-stat__label">磁盘占用</div>
          <div className="storage-stat__value">
            {storageStats ? `${storageStats.diskUsageMb} MB` : "--"}
          </div>
        </div>
        <div className="storage-stat">
          <div className="storage-stat__label">存储上限</div>
          <div className="storage-stat__value">
            {storageStats ? `${storageStats.maxStorageMb} MB` : "--"}
          </div>
        </div>
        <div className="storage-stat">
          <div className="storage-stat__label">活动记录</div>
          <div className="storage-stat__value">
            {storageStats?.activityCount ?? "--"}
          </div>
        </div>
        <div className="storage-stat">
          <div className="storage-stat__label">截图数量</div>
          <div className="storage-stat__value">
            {storageStats?.screenshotCount ?? "--"}
          </div>
        </div>
        <div className="storage-stat">
          <div className="storage-stat__label">保留天数</div>
          <div className="storage-stat__value">
            {storageStats?.retentionDays ?? "--"} 天
          </div>
        </div>
        <div className="storage-stat">
          <div className="storage-stat__label">最早记录</div>
          <div className="storage-stat__value">
            {storageStats?.oldestActivityDate || "--"}
          </div>
        </div>
      </div>

      {storageStats && storageStats.maxStorageMb > 0 && (
        <div className="storage-bar">
          <div className="storage-bar__label">
            使用率: {Math.round((storageStats.diskUsageMb / storageStats.maxStorageMb) * 100)}%
          </div>
          <div className="storage-bar__track">
            <div
              className="storage-bar__fill"
              style={{ width: `${Math.min(100, (storageStats.diskUsageMb / storageStats.maxStorageMb) * 100)}%` }}
            />
          </div>
        </div>
      )}

      <button className="btn" onClick={fetchStorageStats}>
        <RefreshCw size={14} /> 刷新
      </button>
    </div>
  );
}

// ===== 外观设置 =====

function AppearanceSection() {
  const { theme, setTheme } = useThemeStore();

  return (
    <div className="settings-section">
      <h3 className="settings-section__title">主题外观</h3>
      <p className="settings-section__desc">配置 DayLens 客户端的主题模式，修改后立即生效。</p>

      <div className="theme-selector">
        <button
          className={`theme-card ${theme === "light" ? "theme-card--active" : ""}`}
          onClick={() => setTheme("light")}
        >
          <Sun size={24} className="theme-card__icon" />
          <span>浅色</span>
        </button>

        <button
          className={`theme-card ${theme === "dark" ? "theme-card--active" : ""}`}
          onClick={() => setTheme("dark")}
        >
          <Moon size={24} className="theme-card__icon" />
          <span>深色</span>
        </button>

        <button
          className={`theme-card ${theme === "system" ? "theme-card--active" : ""}`}
          onClick={() => setTheme("system")}
        >
          <Monitor size={24} className="theme-card__icon" />
          <span>跟随系统</span>
        </button>
      </div>
    </div>
  );
}

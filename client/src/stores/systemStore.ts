import { create } from "zustand";
import { invoke } from "@tauri-apps/api/core";
import { CMD } from "../utils/api";
import { useLogStore } from "./logStore";
import type { AppConfig, StorageStats } from "../types";

const log = (level: "info" | "warn" | "error" | "success", category: "connection" | "sync" | "capture" | "system", message: string) => {
  useLogStore.getState().addLog(level, category, message);
};

interface SystemState {
  isServerConnected: boolean;
  connectionChecking: boolean;
  syncQueueSize: number;
  storageStats: StorageStats | null;
  config: AppConfig | null;
  configLoaded: boolean;
  fetchConfig: () => Promise<void>;
  fetchSyncQueue: () => Promise<void>;
  fetchStorageStats: () => Promise<void>;
  updateServerUrl: (url: string) => Promise<void>;
  updateServerToken: (token: string) => Promise<void>;
  updateCaptureInterval: (secs: number) => Promise<void>;
  testConnection: () => Promise<boolean>;
  checkConnection: () => Promise<void>;
}

export const useSystemStore = create<SystemState>((set) => ({
  isServerConnected: false,
  connectionChecking: false,
  syncQueueSize: 0,
  storageStats: null,
  config: null,
  configLoaded: false,

  fetchConfig: async () => {
    try {
      const config = await invoke<AppConfig>(CMD.GET_CONFIG);
      set({ config, configLoaded: true });
      log("info", "system", "配置加载完成 — 服务器: " + config.server.url);
    } catch (e) {
      log("error", "system", "配置加载失败: " + e);
    }
  },

  fetchSyncQueue: async () => {
    try {
      const size = await invoke<number>(CMD.GET_SYNC_QUEUE_SIZE);
      const prev = useSystemStore.getState().syncQueueSize;
      set({ syncQueueSize: size });
      if (prev > 0 && size < prev) {
        log("success", "sync", "已同步 " + (prev - size) + " 条，剩余 " + size + " 条");
      }
      if (prev > 0 && size === 0) {
        log("success", "sync", "同步队列已清空 ✓");
      }
      if (size > 0 && size !== prev) {
        log("info", "sync", "同步队列: " + size + " 条待同步");
      }
    } catch {
      // 静默
    }
  },

  fetchStorageStats: async () => {
    try {
      const stats = await invoke<StorageStats>(CMD.GET_STORAGE_STATS);
      set({ storageStats: stats });
    } catch {
      // 静默
    }
  },

  updateServerUrl: async (url: string) => {
    await invoke(CMD.UPDATE_SERVER_URL, { url });
    set((s) => ({
      config: s.config ? { ...s.config, server: { ...s.config.server, url } } : null,
    }));
    log("info", "connection", "服务器地址已更新: " + url);
  },

  updateServerToken: async (token: string) => {
    await invoke(CMD.UPDATE_SERVER_TOKEN, { token });
    set((s) => ({
      config: s.config ? { ...s.config, server: { ...s.config.server, token } } : null,
    }));
    log("info", "connection", "Token 已更新");
  },

  updateCaptureInterval: async (secs: number) => {
    await invoke(CMD.UPDATE_CAPTURE_INTERVAL, { secs });
    set((s) => ({
      config: s.config
        ? { ...s.config, capture: { ...s.config.capture, screenshotIntervalSecs: secs } }
        : null,
    }));
    log("info", "capture", "采集间隔已更新: " + secs + "s");
  },

  testConnection: async () => {
    set({ connectionChecking: true });
    log("info", "connection", "正在测试服务器连接...");
    try {
      const ok = await invoke<boolean>(CMD.TEST_CONNECTION);
      set({ isServerConnected: ok, connectionChecking: false });
      if (ok) {
        log("success", "connection", "服务器连接成功 ✓");
      } else {
        log("error", "connection", "服务器连接失败 — 请检查地址和 Token");
      }
      return ok;
    } catch (e) {
      set({ isServerConnected: false, connectionChecking: false });
      log("error", "connection", "连接测试异常: " + e);
      return false;
    }
  },

  checkConnection: async () => {
    try {
      const ok = await invoke<boolean>(CMD.TEST_CONNECTION);
      const was = useSystemStore.getState().isServerConnected;
      set({ isServerConnected: ok });
      if (ok && !was) {
        log("success", "connection", "服务器连接已恢复");
      } else if (!ok && was) {
        log("error", "connection", "服务器连接已断开");
      } else if (!ok && !was) {
        log("warn", "connection", "服务器仍未连接");
      }
    } catch {
      const was = useSystemStore.getState().isServerConnected;
      set({ isServerConnected: false });
      if (was) {
        log("error", "connection", "服务器连接已断开");
      } else {
        log("warn", "connection", "服务器连接检查失败");
      }
    }
  },
}));

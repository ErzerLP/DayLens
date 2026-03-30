// 全局系统状态 Store

import { create } from "zustand";
import { invoke } from "@tauri-apps/api/core";
import { CMD } from "../utils/api";
import type { AppConfig, StorageStats } from "../types";

interface SystemState {
  isServerAvailable: boolean;
  syncQueueSize: number;
  storageStats: StorageStats | null;
  config: AppConfig | null;
  configLoaded: boolean;

  // Actions
  fetchConfig: () => Promise<void>;
  fetchSyncQueue: () => Promise<void>;
  fetchStorageStats: () => Promise<void>;
  updateServerUrl: (url: string) => Promise<void>;
  updateServerToken: (token: string) => Promise<void>;
  updateCaptureInterval: (secs: number) => Promise<void>;
}

export const useSystemStore = create<SystemState>((set) => ({
  isServerAvailable: false,
  syncQueueSize: 0,
  storageStats: null,
  config: null,
  configLoaded: false,

  fetchConfig: async () => {
    try {
      const config = await invoke<AppConfig>(CMD.GET_CONFIG);
      set({ config, configLoaded: true });
    } catch (e) {
      console.error("加载配置失败:", e);
    }
  },

  fetchSyncQueue: async () => {
    try {
      const size = await invoke<number>(CMD.GET_SYNC_QUEUE_SIZE);
      set({ syncQueueSize: size });
    } catch {
      // 静默失败
    }
  },

  fetchStorageStats: async () => {
    try {
      const stats = await invoke<StorageStats>(CMD.GET_STORAGE_STATS);
      set({ storageStats: stats });
    } catch {
      // 静默失败
    }
  },

  updateServerUrl: async (url: string) => {
    await invoke(CMD.UPDATE_SERVER_URL, { url });
    set((s) => ({
      config: s.config ? { ...s.config, server: { ...s.config.server, url } } : null,
    }));
  },

  updateServerToken: async (token: string) => {
    await invoke(CMD.UPDATE_SERVER_TOKEN, { token });
    set((s) => ({
      config: s.config ? { ...s.config, server: { ...s.config.server, token } } : null,
    }));
  },

  updateCaptureInterval: async (secs: number) => {
    await invoke(CMD.UPDATE_CAPTURE_INTERVAL, { secs });
    set((s) => ({
      config: s.config
        ? {
            ...s.config,
            capture: { ...s.config.capture, screenshotIntervalSecs: secs },
          }
        : null,
    }));
  },
}));

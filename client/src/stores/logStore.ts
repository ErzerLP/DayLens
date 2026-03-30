// 应用日志 Store — 前端内存日志缓冲

import { create } from "zustand";

export type LogLevel = "info" | "warn" | "error" | "success";
export type LogCategory = "connection" | "sync" | "capture" | "system";

export interface LogEntry {
  id: number;
  timestamp: number;
  level: LogLevel;
  category: LogCategory;
  message: string;
}

interface LogState {
  logs: LogEntry[];
  nextId: number;
  maxLogs: number;

  addLog: (level: LogLevel, category: LogCategory, message: string) => void;
  clearLogs: () => void;
}

export const useLogStore = create<LogState>((set) => ({
  logs: [],
  nextId: 1,
  maxLogs: 500,

  addLog: (level, category, message) => {
    set((s) => {
      const entry: LogEntry = {
        id: s.nextId,
        timestamp: Date.now(),
        level,
        category,
        message,
      };
      const logs = [entry, ...s.logs].slice(0, s.maxLogs);
      return { logs, nextId: s.nextId + 1 };
    });
  },

  clearLogs: () => set({ logs: [], nextId: 1 }),
}));

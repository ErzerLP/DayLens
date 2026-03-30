// Tauri invoke 封装 Hook — 带加载/错误状态

import { useState, useCallback } from "react";
import { invoke } from "@tauri-apps/api/core";

interface InvokeState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

/**
 * 通用 invoke Hook
 *
 * 使用方式：
 * ```ts
 * const { data, loading, error, execute } = useInvoke<DailyStats>("get_today_stats");
 * useEffect(() => { execute(); }, [execute]);
 * ```
 */
export function useInvoke<T>(command: string) {
  const [state, setState] = useState<InvokeState<T>>({
    data: null,
    loading: false,
    error: null,
  });

  const execute = useCallback(
    async (args?: Record<string, unknown>) => {
      setState((s) => ({ ...s, loading: true, error: null }));
      try {
        const result = await invoke<T>(command, args);
        setState({ data: result, loading: false, error: null });
        return result;
      } catch (e) {
        const msg = typeof e === "string" ? e : String(e);
        setState((s) => ({ ...s, loading: false, error: msg }));
        return null;
      }
    },
    [command],
  );

  return { ...state, execute };
}

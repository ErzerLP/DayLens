// Tauri 事件监听 Hook

import { useEffect } from "react";
import { listen, type UnlistenFn } from "@tauri-apps/api/event";

/**
 * 监听 Tauri 事件，组件卸载时自动清理。
 */
export function useEventListener<T>(
  event: string,
  handler: (payload: T) => void,
) {
  useEffect(() => {
    let unlisten: UnlistenFn | null = null;

    listen<T>(event, (ev) => handler(ev.payload)).then((fn) => {
      unlisten = fn;
    });

    return () => {
      unlisten?.();
    };
  }, [event, handler]);
}

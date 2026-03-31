import { create } from "zustand";
import { persist } from "zustand/middleware";

export type ThemeMode = "light" | "dark" | "system";

interface ThemeState {
  theme: ThemeMode;
  setTheme: (theme: ThemeMode) => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme: "system",
      setTheme: (theme) => set({ theme }),
    }),
    {
      name: "daylens-theme",
    }
  )
);

/**
 * 初始化并将当前活跃的主题应用到 <html> 元素上。
 * 建议在 App 或 main.tsx 挂载阶段执行一次，并在事件中持续监听。
 */
export function syncThemeToDom(theme: ThemeMode) {
  const isDark =
    theme === "dark" ||
    (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches);

  const root = document.documentElement;

  if (isDark) {
    root.classList.add("dark");
    root.classList.remove("light");
  } else {
    root.classList.add("light");
    root.classList.remove("dark");
  }
}

/**
 * 设置系统偏好监听器（当 theme="system" 时，随系统直接切换）
 */
export function setupSystemThemeListener() {
  const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");

  const handleChange = () => {
    // 只有当设置为系统跟随主题时，才实时应用
    const currentTheme = useThemeStore.getState().theme;
    if (currentTheme === "system") {
      syncThemeToDom("system");
    }
  };

  // 现代 API (旧版本可能需要 addListener)
  if (mediaQuery.addEventListener) {
    mediaQuery.addEventListener("change", handleChange);
  } else {
    mediaQuery.addListener(handleChange);
  }

  return () => {
    if (mediaQuery.removeEventListener) {
      mediaQuery.removeEventListener("change", handleChange);
    } else {
      mediaQuery.removeListener(handleChange);
    }
  };
}

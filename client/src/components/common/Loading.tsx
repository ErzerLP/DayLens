import { Loader2 } from "lucide-react";
import "./Loading.css";

interface Props {
  text?: string;
  size?: "sm" | "md" | "lg";
}

/** 居中加载指示器 */
export function LoadingSpinner({ text = "加载中…", size = "md" }: Props) {
  const iconSize = { sm: 16, md: 24, lg: 36 }[size];

  return (
    <div className={`loading-spinner loading-spinner--${size}`}>
      <Loader2 size={iconSize} className="loading-spinner__icon" />
      {text && <span className="loading-spinner__text">{text}</span>}
    </div>
  );
}

/** 行内小加载 */
export function InlineLoader({ text }: { text?: string }) {
  return (
    <span className="inline-loader">
      <Loader2 size={14} className="loading-spinner__icon" />
      {text}
    </span>
  );
}

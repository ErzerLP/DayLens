import type { LucideIcon } from "lucide-react";
import { Inbox } from "lucide-react";
import "./EmptyState.css";

interface Props {
  icon?: LucideIcon;
  title: string;
  description?: string;
}

/** 通用空状态展示 */
export function EmptyState({ icon: Icon = Inbox, title, description }: Props) {
  return (
    <div className="empty-state">
      <Icon size={48} strokeWidth={1} className="empty-state__icon" />
      <div className="empty-state__title">{title}</div>
      {description && (
        <div className="empty-state__desc">{description}</div>
      )}
    </div>
  );
}

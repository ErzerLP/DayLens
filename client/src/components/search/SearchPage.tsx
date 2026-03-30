import { useState, useRef, useEffect } from "react";
import { motion } from "framer-motion";
import { Search, Send, Bot, User, Sparkles } from "lucide-react";
import { invoke } from "@tauri-apps/api/core";
import ReactMarkdown from "react-markdown";
import { CMD } from "../../utils/api";
import { formatTime } from "../../utils/format";
import type { SearchResultItem, ChatMessage, AssistantReply } from "../../types";
import "./SearchAI.css";

type Tab = "search" | "chat";

export default function SearchPage() {
  const [tab, setTab] = useState<Tab>("search");

  return (
    <div className="search-page">
      <div className="search-page__tabs">
        <button
          className={`tab-btn ${tab === "search" ? "tab-btn--active" : ""}`}
          onClick={() => setTab("search")}
        >
          <Search size={14} /> 搜索
        </button>
        <button
          className={`tab-btn ${tab === "chat" ? "tab-btn--active" : ""}`}
          onClick={() => setTab("chat")}
        >
          <Sparkles size={14} /> AI 助手
        </button>
      </div>

      {tab === "search" ? <SearchPanel /> : <ChatPanel />}
    </div>
  );
}

// ===== 搜索面板 =====

function SearchPanel() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResultItem[]>([]);
  const [loading, setLoading] = useState(false);

  const doSearch = async () => {
    if (!query.trim()) return;
    setLoading(true);
    try {
      const res = await invoke<SearchResultItem[]>(CMD.SEARCH_ACTIVITIES, {
        query: query.trim(),
        limit: 30,
      });
      setResults(res);
    } catch (e) {
      console.error("搜索失败:", e);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="search-panel">
      <div className="search-panel__input-row">
        <Search size={16} className="search-panel__icon" />
        <input
          className="search-panel__input"
          placeholder="搜索活动标题、OCR 文本…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && doSearch()}
        />
        <button className="btn btn--accent" onClick={doSearch} disabled={loading}>
          搜索
        </button>
      </div>

      <div className="search-results">
        {results.map((r) => (
          <motion.div
            key={r.activityId}
            className="search-result-item"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
          >
            <div className="search-result-item__header">
              <span className="search-result-item__app">{r.appName}</span>
              <span className="search-result-item__time">{formatTime(r.timestamp)}</span>
            </div>
            <div className="search-result-item__title">{r.windowTitle}</div>
            {r.matchedText && (
              <div className="search-result-item__match">{r.matchedText}</div>
            )}
          </motion.div>
        ))}
        {!loading && results.length === 0 && query && (
          <div className="search-results__empty">无匹配结果</div>
        )}
      </div>
    </div>
  );
}

// ===== AI 对话面板 =====

interface DisplayMessage {
  role: "user" | "assistant";
  content: string;
}

function ChatPanel() {
  const [messages, setMessages] = useState<DisplayMessage[]>([]);
  const [input, setInput] = useState("");
  const [thinking, setThinking] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: "smooth" });
  }, [messages]);

  const sendMessage = async () => {
    if (!input.trim() || thinking) return;

    const userMsg: DisplayMessage = { role: "user", content: input.trim() };
    setMessages((prev) => [...prev, userMsg]);
    setInput("");
    setThinking(true);

    try {
      const chatHistory: ChatMessage[] = [...messages, userMsg].map((m) => ({
        role: m.role,
        content: m.content,
      }));

      const reply = await invoke<AssistantReply>(CMD.CHAT_AI, {
        messages: chatHistory,
        tools: [],
      });

      setMessages((prev) => [...prev, { role: "assistant", content: reply.content }]);
    } catch (e) {
      setMessages((prev) => [
        ...prev,
        { role: "assistant", content: `错误：${String(e)}` },
      ]);
    } finally {
      setThinking(false);
    }
  };

  return (
    <div className="chat-panel">
      <div className="chat-panel__messages" ref={scrollRef}>
        {messages.length === 0 && (
          <div className="chat-panel__welcome">
            <Bot size={36} strokeWidth={1} />
            <p>Hi! 我是你的工作分析助手。</p>
            <p className="chat-panel__hint">你可以问我关于今天工作模式、效率分析等问题。</p>
          </div>
        )}

        {messages.map((msg, i) => (
          <motion.div
            key={i}
            className={`chat-bubble chat-bubble--${msg.role}`}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.2 }}
          >
            <div className="chat-bubble__avatar">
              {msg.role === "user" ? <User size={14} /> : <Bot size={14} />}
            </div>
            <div className="chat-bubble__content">
              {msg.role === "assistant" ? (
                <ReactMarkdown>{msg.content}</ReactMarkdown>
              ) : (
                msg.content
              )}
            </div>
          </motion.div>
        ))}

        {thinking && (
          <div className="chat-bubble chat-bubble--assistant">
            <div className="chat-bubble__avatar"><Bot size={14} /></div>
            <div className="chat-thinking">
              <span /><span /><span />
            </div>
          </div>
        )}
      </div>

      <div className="chat-panel__input-row">
        <input
          className="chat-panel__input"
          placeholder="输入问题…"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && sendMessage()}
          disabled={thinking}
        />
        <button
          className="icon-btn chat-panel__send"
          onClick={sendMessage}
          disabled={thinking || !input.trim()}
        >
          <Send size={16} />
        </button>
      </div>
    </div>
  );
}

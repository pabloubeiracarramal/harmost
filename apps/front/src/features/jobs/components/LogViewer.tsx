import { useEffect, useRef, useState } from 'react';
import type { LogLine } from '@/shared/ws/wsClient';

const AT_BOTTOM_THRESHOLD_PX = 40;

export function LogViewer({ lines, live }: { lines: LogLine[]; live: boolean }) {
  const containerRef = useRef<HTMLDivElement>(null);
  // Pinned = follow new output. Scrolling up pauses; scrolling back down resumes.
  const [pinned, setPinned] = useState(true);

  useEffect(() => {
    const el = containerRef.current;
    if (el && pinned) el.scrollTop = el.scrollHeight;
  }, [lines, pinned]);

  const handleScroll = () => {
    const el = containerRef.current;
    if (!el) return;
    setPinned(el.scrollHeight - el.scrollTop - el.clientHeight < AT_BOTTOM_THRESHOLD_PX);
  };

  const jumpToBottom = () => {
    const el = containerRef.current;
    if (el) el.scrollTop = el.scrollHeight;
    setPinned(true);
  };

  return (
    <div className="overflow-hidden rounded-xl border border-neutral-800">
      <div className="flex items-center justify-between border-b border-neutral-800 bg-neutral-900 px-4 py-2.5">
        <span className="text-xs font-medium uppercase tracking-widest text-neutral-500">
          Logs
        </span>
        <span className="flex items-center gap-3">
          {live && (
            <span className="flex items-center gap-1.5 text-xs text-neutral-500">
              <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-emerald-400" />
              live
            </span>
          )}
          {!pinned && (
            <button
              onClick={jumpToBottom}
              className="rounded-md bg-neutral-800 px-2 py-1 text-xs text-neutral-300 transition hover:bg-neutral-700"
            >
              ↓ Jump to bottom
            </button>
          )}
        </span>
      </div>
      <div
        ref={containerRef}
        onScroll={handleScroll}
        className="max-h-[28rem] overflow-y-auto bg-neutral-950 p-4 font-mono text-xs leading-5"
      >
        {lines.length === 0 ? (
          <p className="text-neutral-600">No output yet.</p>
        ) : (
          lines.map((l) => (
            <div key={l.sequence} className="flex gap-3 whitespace-pre-wrap break-all">
              <span className="shrink-0 select-none text-neutral-700">
                {String(l.sequence).padStart(4, ' ')}
              </span>
              <span className={l.stream === 'stderr' ? 'text-red-400' : 'text-neutral-200'}>
                {l.line}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

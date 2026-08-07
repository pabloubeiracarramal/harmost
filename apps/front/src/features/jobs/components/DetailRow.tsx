export function DetailRow({ label, value, link }: { label: string; value: string; link?: string }) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-neutral-800 py-2.5 last:border-0">
      <span className="shrink-0 text-sm text-neutral-400">{label}</span>
      {link ? (
        <a href={link} className="truncate font-mono text-sm text-indigo-400 hover:text-indigo-300">
          {value}
        </a>
      ) : (
        <span className="truncate font-mono text-sm text-white">{value}</span>
      )}
    </div>
  );
}

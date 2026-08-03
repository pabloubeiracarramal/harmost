export function Field({
  label,
  hint,
  required,
  children,
}: {
  label: string;
  hint?: string;
  required?: boolean;
  children: React.ReactNode;
}) {
  return (
    <label className="block">
      <span className="mb-1.5 block text-sm font-medium text-neutral-300">
        {label}
        {required && <span className="text-red-400"> *</span>}
      </span>
      {children}
      {hint && <span className="mt-1 block text-xs text-neutral-600">{hint}</span>}
    </label>
  );
}

export function TypeBadge({ type }: { type: string }) {
  if (!type) return null;
  return (
    <span className="inline-flex items-center rounded-full bg-gray-50 px-2 py-0.5 text-xs font-medium text-gray-500 ring-1 ring-gray-200 ring-inset">
      {type}
    </span>
  );
}

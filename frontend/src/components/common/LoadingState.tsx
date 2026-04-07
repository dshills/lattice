export function LoadingState() {
  return (
    <div className="animate-pulse space-y-4 py-4" role="status" aria-busy="true">
      <div className="h-4 w-48 rounded bg-gray-200" />
      <div className="h-4 w-64 rounded bg-gray-200" />
      <div className="h-4 w-40 rounded bg-gray-200" />
      <div className="h-32 rounded bg-gray-100" />
    </div>
  );
}

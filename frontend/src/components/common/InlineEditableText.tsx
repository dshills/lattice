import { useState, useRef, useEffect, useCallback } from "react";

interface InlineEditableTextProps {
  value: string;
  onSave: (value: string) => Promise<void>;
  as?: "input" | "textarea";
  className?: string;
  placeholder?: string;
}

export function InlineEditableText({
  value,
  onSave,
  as = "input",
  className = "",
  placeholder,
}: InlineEditableTextProps) {
  const [draft, setDraft] = useState(value);
  const [status, setStatus] = useState<"idle" | "saving" | "saved" | "error">(
    "idle",
  );
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const latestDraft = useRef(draft);
  const focusedRef = useRef(false);
  const savingRef = useRef(false);

  // Only sync prop to draft when not focused and not saving
  useEffect(() => {
    if (!focusedRef.current && !savingRef.current) {
      setDraft(value);
    }
  }, [value]);

  useEffect(() => {
    latestDraft.current = draft;
  }, [draft]);

  // Cleanup timer on unmount
  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  const save = useCallback(
    async (val: string) => {
      if (val === value) return;
      savingRef.current = true;
      setStatus("saving");
      try {
        await onSave(val);
        setStatus("saved");
        savingRef.current = false;
        setTimeout(() => setStatus("idle"), 1500);
      } catch {
        savingRef.current = false;
        setStatus("error");
      }
    },
    [value, onSave],
  );

  const handleChange = (val: string) => {
    setDraft(val);
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      save(val);
    }, 500);
  };

  const handleBlur = () => {
    focusedRef.current = false;
    if (timerRef.current) clearTimeout(timerRef.current);
    save(latestDraft.current);
  };

  const handleFocus = () => {
    focusedRef.current = true;
  };

  const statusIndicator =
    status === "saving" ? (
      <span className="text-xs text-gray-400">Saving...</span>
    ) : status === "saved" ? (
      <span className="text-xs text-green-500">Saved</span>
    ) : status === "error" ? (
      <span className="text-xs text-red-500">Failed to save</span>
    ) : null;

  const baseClasses = `w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-1 ${
    status === "error"
      ? "border-red-300 focus:border-red-500 focus:ring-red-500"
      : "border-gray-300 focus:border-blue-500 focus:ring-blue-500"
  } ${className}`;

  return (
    <div>
      {as === "textarea" ? (
        <textarea
          value={draft}
          onChange={(e) => handleChange(e.target.value)}
          onBlur={handleBlur}
          onFocus={handleFocus}
          className={`${baseClasses} min-h-[100px] resize-y`}
          placeholder={placeholder}
        />
      ) : (
        <input
          type="text"
          value={draft}
          onChange={(e) => handleChange(e.target.value)}
          onBlur={handleBlur}
          onFocus={handleFocus}
          className={baseClasses}
          placeholder={placeholder}
        />
      )}
      <div className="mt-1 h-4">{statusIndicator}</div>
    </div>
  );
}

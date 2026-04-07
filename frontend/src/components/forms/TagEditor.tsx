import { useState } from "react";
import { TagBadge } from "../workitems/TagBadge";

interface TagEditorProps {
  tags: string[];
  onUpdate: (tags: string[]) => void;
}

export function TagEditor({ tags, onUpdate }: TagEditorProps) {
  const [input, setInput] = useState("");

  const addTag = () => {
    const trimmed = input.trim().toLowerCase();
    if (!trimmed || tags.includes(trimmed)) return;
    onUpdate([...tags, trimmed]);
    setInput("");
  };

  const removeTag = (tag: string) => {
    onUpdate(tags.filter((t) => t !== tag));
  };

  return (
    <div>
      <div className="mb-2 flex flex-wrap gap-1">
        {tags.map((tag) => (
          <button
            key={tag}
            onClick={() => removeTag(tag)}
            className="group inline-flex items-center gap-0.5"
            aria-label={`Remove tag ${tag}`}
          >
            <TagBadge tag={tag} />
            <span className="text-xs text-gray-400 group-hover:text-red-500">
              &times;
            </span>
          </button>
        ))}
      </div>
      <input
        type="text"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            e.preventDefault();
            addTag();
          }
        }}
        placeholder="Add tag..."
        className="w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    </div>
  );
}

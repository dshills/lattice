interface UserInitialsProps {
  name: string;
}

export function UserInitials({ name }: UserInitialsProps) {
  const initials = name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  return (
    <span
      className="flex h-5 w-5 items-center justify-center rounded-full bg-blue-100 text-[10px] font-medium text-blue-700"
      title={name}
    >
      {initials}
    </span>
  );
}

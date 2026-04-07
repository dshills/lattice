import { useParams } from "react-router";

export function ItemDetailPage() {
  const { id } = useParams<{ id: string }>();
  return <h1 className="text-2xl font-semibold">Item {id}</h1>;
}

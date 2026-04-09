import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "react-router";
import {
  ReactFlow,
  Background,
  Controls,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
  MarkerType,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import dagre from "@dagrejs/dagre";
import { useWorkItems, useWorkItem } from "../hooks/useWorkItems";
import { useCycles } from "../hooks/useCycles";
import { getWorkItem } from "../lib/api/workitems";
import { GraphNode } from "../components/graph/GraphNode";
import { GraphDetailPanel } from "../components/graph/GraphDetailPanel";
import { LoadingState } from "../components/common/LoadingState";
import type { WorkItem, RelationshipType } from "../lib/types";

const nodeTypes = { workItem: GraphNode };

const EDGE_STYLES: Record<
  RelationshipType,
  { strokeDasharray?: string; stroke: string }
> = {
  depends_on: { stroke: "#6B7280" },
  blocks: { stroke: "#EF4444" },
  relates_to: { strokeDasharray: "5 5", stroke: "#9CA3AF" },
  duplicate_of: { strokeDasharray: "2 2", stroke: "#D1D5DB" },
};

function layoutGraph(nodes: Node[], edges: Edge[]): Node[] {
  const g = new dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({ rankdir: "TB", nodesep: 50, ranksep: 80 });

  for (const node of nodes) {
    g.setNode(node.id, { width: 180, height: 80 });
  }
  for (const edge of edges) {
    g.setEdge(edge.source, edge.target);
  }

  dagre.layout(g);

  return nodes.map((node) => {
    const pos = g.node(node.id);
    return {
      ...node,
      position: { x: pos.x - 90, y: pos.y - 40 },
    };
  });
}

export function GraphPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const focusId = searchParams.get("focus") ?? "";
  const [selectedItem, setSelectedItem] = useState<WorkItem | null>(null);
  const [neighborItems, setNeighborItems] = useState<Map<string, WorkItem>>(
    new Map(),
  );
  const fetchIdRef = useRef(0);

  const { data: focusItem } = useWorkItem(focusId);
  const { data: cycleData } = useCycles(focusId, !!focusId);

  const cycleEdges = useMemo(() => {
    if (!cycleData?.has_cycle || !cycleData.cycle) return new Set<string>();
    const edges = new Set<string>();
    for (let i = 0; i < cycleData.cycle.length - 1; i++) {
      edges.add(`${cycleData.cycle[i]}-${cycleData.cycle[i + 1]}`);
    }
    return edges;
  }, [cycleData]);

  // Load neighbors (1 hop) with stale request protection
  useEffect(() => {
    if (!focusItem) return;
    const fetchId = ++fetchIdRef.current;
    const ids = focusItem.relationships.map((r) => r.target_id);
    const unique = [...new Set(ids)];

    Promise.allSettled(unique.map((id) => getWorkItem(id))).then((results) => {
      if (fetchId !== fetchIdRef.current) return;
      const map = new Map<string, WorkItem>();
      for (const result of results) {
        if (result.status === "fulfilled") {
          map.set(result.value.id, result.value);
        }
      }
      setNeighborItems(map);
    });
  }, [focusItem]);

  // Build graph topology (layout only when structure changes)
  const { layoutNodes, graphEdges } = useMemo(() => {
    if (!focusItem) return { layoutNodes: [], graphEdges: [] };

    const items = new Map<string, WorkItem>([[focusItem.id, focusItem]]);
    for (const [id, item] of neighborItems) {
      items.set(id, item);
    }

    const nodes: Node[] = Array.from(items.values()).map((item) => ({
      id: item.id,
      type: "workItem",
      position: { x: 0, y: 0 },
      data: { item, selected: false },
    }));

    const edges: Edge[] = [];
    for (const item of items.values()) {
      for (const rel of item.relationships) {
        if (!items.has(rel.target_id)) continue;
        const style = EDGE_STYLES[rel.type] ?? EDGE_STYLES.relates_to;
        const isCycleEdge = cycleEdges.has(`${item.id}-${rel.target_id}`);

        edges.push({
          id: rel.id,
          source: item.id,
          target: rel.target_id,
          label: rel.type.replace("_", " "),
          style: isCycleEdge
            ? { stroke: "#EF4444", strokeWidth: 2, strokeDasharray: "8 4" }
            : { ...style, strokeWidth: 1.5 },
          animated: isCycleEdge,
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: isCycleEdge ? "#EF4444" : style.stroke,
          },
        });
      }
    }

    const laidOut = layoutGraph(nodes, edges);
    return { layoutNodes: laidOut, graphEdges: edges };
  }, [focusItem, neighborItems, cycleEdges]);

  const [nodes, setNodes, onNodesChange] = useNodesState(layoutNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(graphEdges);

  // Apply layout only when topology changes
  useEffect(() => {
    setNodes(layoutNodes);
    setEdges(graphEdges);
  }, [layoutNodes, graphEdges, setNodes, setEdges]);

  // Update selection highlighting without re-layout
  useEffect(() => {
    setNodes((nds: Node[]) =>
      nds.map((n: Node) => ({
        ...n,
        data: { ...n.data, selected: selectedItem?.id === n.id },
      })),
    );
  }, [selectedItem, setNodes]);

  const handleNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      const item =
        node.id === focusItem?.id
          ? focusItem
          : neighborItems.get(node.id) ?? null;
      setSelectedItem(item);
    },
    [focusItem, neighborItems],
  );

  const handleFocus = useCallback(
    (id: string) => {
      setSearchParams({ focus: id });
      setSelectedItem(null);
    },
    [setSearchParams],
  );

  // Auto-focus: pick the first available item when no focus param is set
  const { data: allItems, isLoading: loadingAll } = useWorkItems({
    page_size: 1,
  });

  useEffect(() => {
    if (!focusId && allItems?.items?.length) {
      setSearchParams({ focus: allItems.items[0].id }, { replace: true });
    }
  }, [focusId, allItems, setSearchParams]);

  if (!focusId) {
    if (loadingAll) return <LoadingState />;
    return (
      <div className="flex flex-col items-center justify-center gap-4 py-20">
        <h1 className="text-xl font-semibold text-gray-700">
          Dependency Graph
        </h1>
        <p className="text-gray-500">No work items to display</p>
      </div>
    );
  }

  return (
    <div className="flex -m-6 flex-1" style={{ height: "calc(100vh - var(--header-height, 3.5rem))" }}>
      {cycleData?.has_cycle && (
        <div className="absolute top-2 left-1/2 -translate-x-1/2 z-10 rounded-md bg-amber-50 border border-amber-200 px-4 py-2 text-sm text-amber-800">
          Dependency cycle detected
        </div>
      )}

      <div className="flex-1">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeClick={handleNodeClick}
          nodeTypes={nodeTypes}
          fitView
          minZoom={0.2}
          maxZoom={2}
        >
          <Background />
          <Controls />
        </ReactFlow>
      </div>

      {selectedItem && (
        <GraphDetailPanel item={selectedItem} onFocus={handleFocus} />
      )}
    </div>
  );
}

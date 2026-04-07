package graph

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dshills/lattice/internal/domain"
)

// CycleDetector detects dependency graph cycles using DFS.
// It follows depends_on and blocks edges as directed A→B.
type CycleDetector struct {
	db *sql.DB
}

// NewCycleDetector creates a CycleDetector backed by the given database.
func NewCycleDetector(db *sql.DB) *CycleDetector {
	return &CycleDetector{db: db}
}

// DetectCycles finds all unique cycles that include the specified WorkItem.
// It follows depends_on and blocks edges (both treated as directed source→target).
// Each returned cycle is a list of WorkItem IDs in traversal order.
func (d *CycleDetector) DetectCycles(ctx context.Context, workItemID string) ([][]string, error) {
	// Build the adjacency list for the relevant subgraph by loading all
	// depends_on and blocks edges reachable from workItemID.
	adj, err := d.buildSubgraph(ctx, workItemID)
	if err != nil {
		return nil, err
	}

	const maxDepth = 200 // bound search to prevent runaway traversal

	var cycles [][]string
	seen := make(map[string]bool)

	// DFS to find all cycles that include workItemID.
	var dfs func(node string, path []string)
	dfs = func(node string, path []string) {
		if ctx.Err() != nil {
			return
		}
		if len(path) > maxDepth {
			return
		}
		for _, neighbor := range adj[node] {
			if neighbor == workItemID {
				// Found a cycle back to the start.
				cycle := make([]string, len(path))
				copy(cycle, path)
				cycles = append(cycles, cycle)
				continue
			}
			if seen[neighbor] {
				continue
			}
			seen[neighbor] = true
			dfs(neighbor, append(path, neighbor))
			seen[neighbor] = false
		}
	}

	seen[workItemID] = true
	dfs(workItemID, []string{workItemID})

	return dedupCycles(cycles), nil
}

// buildSubgraph loads all depends_on and blocks edges reachable from startID
// using a recursive CTE (single query), and returns an adjacency list (source → [targets]).
func (d *CycleDetector) buildSubgraph(ctx context.Context, startID string) (map[string][]string, error) {
	const query = `
		WITH RECURSIVE reachable AS (
			SELECT source_id, target_id
			FROM work_item_relationships
			WHERE source_id = ? AND type IN (?, ?)
			UNION
			SELECT r.source_id, r.target_id
			FROM work_item_relationships r
			JOIN reachable rch ON r.source_id = rch.target_id
			WHERE r.type IN (?, ?)
		)
		SELECT source_id, target_id FROM reachable`

	rows, err := d.db.QueryContext(ctx, query,
		startID,
		string(domain.DependsOn), string(domain.Blocks),
		string(domain.DependsOn), string(domain.Blocks),
	)
	if err != nil {
		return nil, fmt.Errorf("load subgraph: %w", err)
	}
	defer func() { _ = rows.Close() }()

	adj := make(map[string][]string)
	for rows.Next() {
		var src, tgt string
		if err := rows.Scan(&src, &tgt); err != nil {
			return nil, fmt.Errorf("scan edge: %w", err)
		}
		adj[src] = append(adj[src], tgt)
	}
	return adj, rows.Err()
}

// dedupCycles removes duplicate cycles. Two cycles are considered equal if they
// contain the same set of nodes in the same cyclic order.
func dedupCycles(cycles [][]string) [][]string {
	if len(cycles) == 0 {
		return cycles
	}

	seen := make(map[string]bool)
	var result [][]string

	for _, cycle := range cycles {
		key := canonicalKey(cycle)
		if !seen[key] {
			seen[key] = true
			result = append(result, cycle)
		}
	}
	return result
}

// canonicalKey produces a canonical string for a cycle by rotating it so the
// lexicographically smallest element comes first.
func canonicalKey(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}

	// Find the index of the lexicographically smallest element.
	minIdx := 0
	for i := 1; i < len(cycle); i++ {
		if cycle[i] < cycle[minIdx] {
			minIdx = i
		}
	}

	// Build the canonical key by rotating.
	var b strings.Builder
	for i := range cycle {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(cycle[(minIdx+i)%len(cycle)])
	}
	return b.String()
}

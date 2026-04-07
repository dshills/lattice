# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Lattice** (working name: Atlas) is a lightweight, graph-based work tracking system. It replaces tools like Jira with a minimal interface: only three states (NotDone, InProgress, Completed), free-form tags, and first-class relationships between work items. Complexity lives in the data model and views, never in user-facing workflows.

The spec lives at `specs/initial/SPEC.md`.

## Technology

- **Language:** Go
- **Database:** MySQL (three tables: `work_items`, `work_item_tags`, `work_item_relationships`)
- **API:** REST (endpoints under `/workitems`)

## Build & Test Commands

```bash
go build ./...
go test ./...
go test -run TestName ./path/to/package   # single test
golangci-lint run ./...                    # lint after any Go changes
```

## Development Workflow

Specs and plans live in `./specs/`. Follow the standard workflow:

1. `/spec-review` — validate SPEC.md with speccritic
2. `/plan` — generate PLAN.md from spec, validate with plancritic
3. `/implement` — implement one phase at a time
4. `/phase-review` — prism review, realitycheck, clarion pack, verifier analyze
5. `/commit` when a phase passes all validation

## Key Domain Rules

- **State transitions are forward-only:** NotDone -> InProgress -> Completed (override flag for admin)
- **Tags must never affect state transitions** — they are metadata only
- **Relationships are directional** with reverse lookup support
- **Only one parent per WorkItem** but arbitrary depth
- **A WorkItem is "blocked"** if it has a `depends_on` where the target is not Completed
- **Circular dependencies** are allowed but must be detectable

## Atlas Index

This repository has an Atlas index for structural and semantic code queries.
Use atlas commands with --agent for compact JSON instead of reading source files:

- `atlas find symbol <name> --agent` — find symbol definitions
- `atlas who-calls <symbol> --agent` — find callers
- `atlas calls <symbol> --agent` — find callees
- `atlas implementations <interface> --agent` — find implementations
- `atlas tests-for <symbol> --agent` — find related tests
- `atlas summarize file <path> --agent` — get file summary
- `atlas list routes --agent` — list HTTP routes
- `atlas export graph --agent` — get full dependency graph

The index auto-updates via a PreToolUse hook. To manually re-index: `atlas index`

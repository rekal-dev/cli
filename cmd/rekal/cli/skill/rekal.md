---
name: rekal
description: |
  Use this skill when working in a repo with Rekal initialized (.rekal/ exists).
  Rekal gives you memory of prior AI sessions — who changed what, why, and when.
  Use `rekal query` to search conversation history, tool calls, and file changes
  before making decisions about code you're modifying.
---

# Rekal — Session Memory

Rekal captures AI coding sessions (conversation turns, tool calls, file changes) and stores them in a local DuckDB database. Use it to understand prior context before modifying code.

## When to Use

- Before modifying a file — check what prior sessions touched it
- When you need context about why code looks the way it does
- When the user asks about prior session history
- When working on files that were recently changed by AI agents

## How to Query

Run SQL against the data DB:

```bash
rekal query "SELECT ..."
```

Read-only (SELECT only). Output is one JSON object per row.

## Schema

### `sessions` — one row per captured session

| Column | Type | Description |
|--------|------|-------------|
| `id` | VARCHAR PK | ULID |
| `parent_session_id` | VARCHAR | Parent session (for subagents) |
| `session_hash` | VARCHAR | SHA-256 of raw session file |
| `captured_at` | TIMESTAMP | When captured |
| `actor_type` | VARCHAR | `"human"` or `"agent"` |
| `agent_id` | VARCHAR | Agent ID if actor_type is agent |
| `user_email` | VARCHAR | Git user.email |
| `branch` | VARCHAR | Git branch |

### `turns` — conversation turns

| Column | Type | Description |
|--------|------|-------------|
| `id` | VARCHAR PK | ULID |
| `session_id` | VARCHAR FK | → sessions.id |
| `turn_index` | INTEGER | 0-based position |
| `role` | VARCHAR | `"human"` (prompt) or `"assistant"` (response) |
| `content` | VARCHAR | Text content |
| `ts` | TIMESTAMP | Timestamp |

### `tool_calls` — tool invocations

| Column | Type | Description |
|--------|------|-------------|
| `id` | VARCHAR PK | ULID |
| `session_id` | VARCHAR FK | → sessions.id |
| `call_order` | INTEGER | 0-based position |
| `tool` | VARCHAR | Write, Edit, Read, Bash, Glob, Grep, Task, etc. |
| `path` | VARCHAR | File path (if applicable) |
| `cmd_prefix` | VARCHAR | First 100 chars of Bash command |

### `checkpoints` — linked to git commits

| Column | Type | Description |
|--------|------|-------------|
| `id` | VARCHAR PK | Orphan branch commit SHA |
| `git_sha` | VARCHAR | Main repo HEAD at checkpoint |
| `git_branch` | VARCHAR | Main repo branch |
| `user_email` | VARCHAR | Git user.email |
| `ts` | TIMESTAMP | Checkpoint time |

### `files_touched` — files changed per checkpoint

| Column | Type | Description |
|--------|------|-------------|
| `id` | VARCHAR PK | ULID |
| `checkpoint_id` | VARCHAR FK | → checkpoints.id |
| `file_path` | VARCHAR | Relative path |
| `change_type` | VARCHAR | A (added), M (modified), D (deleted) |

### `checkpoint_sessions` — junction

| Column | Type |
|--------|------|
| `checkpoint_id` | VARCHAR FK → checkpoints.id |
| `session_id` | VARCHAR FK → sessions.id |

## Common Queries

**What sessions touched a file:**
```bash
rekal query "SELECT DISTINCT t.session_id, s.user_email, s.captured_at FROM tool_calls t JOIN sessions s ON t.session_id = s.id WHERE t.path LIKE '%auth%' ORDER BY s.captured_at DESC"
```

**What was discussed about a file:**
```bash
rekal query "SELECT tu.role, tu.content FROM turns tu JOIN tool_calls tc ON tu.session_id = tc.session_id WHERE tc.path LIKE '%login.tsx%' AND tu.role = 'human' ORDER BY tu.turn_index"
```

**Recent sessions:**
```bash
rekal query "SELECT id, user_email, branch, captured_at FROM sessions ORDER BY captured_at DESC LIMIT 5"
```

**What tools were used most:**
```bash
rekal query "SELECT tool, count(*) as cnt FROM tool_calls GROUP BY tool ORDER BY cnt DESC"
```

**Files most frequently edited by AI:**
```bash
rekal query "SELECT path, count(*) as cnt FROM tool_calls WHERE tool IN ('Write', 'Edit') AND path IS NOT NULL GROUP BY path ORDER BY cnt DESC LIMIT 10"
```

## Guidelines

- Query before modifying files that have prior session history
- Use `LIKE '%filename%'` for fuzzy file matching
- Join `turns` with `tool_calls` via `session_id` to get context around file changes
- Human turns contain the intent; assistant turns contain the reasoning
- `actor_type` distinguishes human sessions from automated agent sessions

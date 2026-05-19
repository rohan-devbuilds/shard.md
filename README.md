# Shard

Make your AI sessions last longer with context compression.

Shard is an open-source Go CLI that reduces AI coding token usage by compressing long conversations into categorized Markdown memory files. It stores durable project knowledge in `.shard/`, retrieves only relevant context before each prompt, and periodically summarizes the active session so useful details survive without resending everything.

## Quick Demo

```bash
shard init
shard chat
```

Inside chat:

```text
> help me fix the Anthropic provider error

Model: claude-3-5-sonnet-20241022 | Effort: medium | Context: 108 tokens

The provider is reading ANTHROPIC_API_KEY correctly. The next thing to check is the request payload and response error body.

> /optimize
Optimized memory. Before: 1842 tokens | After: 94 tokens
```

## Why Shard

AI coding sessions get expensive and brittle when every turn carries old logs, repeated reasoning, and stale context. Shard keeps the durable parts and drops the noise.

- Save tokens by sending compact memory instead of full chat history.
- Run longer sessions by periodically compressing active conversation state.
- Keep project memory readable in plain Markdown.
- Avoid loading unrelated context by using simple keyword retrieval.

## Install

Install from the repository:

```bash
git clone <repo-url>
cd shard.md
go install .
```

Or run without installing:

```bash
go run . init
go run . chat
```

Shard uses Cobra for CLI commands and currently supports Anthropic as the first provider.

## Setup

Set `ANTHROPIC_API_KEY` before using `shard chat` or `shard optimize`.

macOS/Linux:

```bash
export ANTHROPIC_API_KEY="your-api-key"
```

Windows PowerShell:

```powershell
$env:ANTHROPIC_API_KEY="your-api-key"
```

Choose a model with `--model`:

```bash
shard --model claude-3-5-sonnet-20241022 chat
```

## Commands

```bash
shard init
```

Creates `.shard/` in the current directory and initializes Markdown memory files:

- `current.md`
- `tasks.md`
- `bugs.md`
- `architecture.md`
- `decisions.md`
- `codebase.md`
- `commands.md`
- `dependencies.md`
- `api.md`
- `ui.md`
- `agents.md`
- `changelog.md`

Each file starts with frontmatter:

```yaml
---
category: current
updated: 2026-05-19T12:00:00Z
priority: medium
related: []
---
```

```bash
shard chat
```

Starts the interactive chat loop.

```bash
shard context "fix provider error"
```

Shows which memory files would be loaded for a prompt and the estimated context size.

```bash
shard stats
```

Shows session statistics. In standalone mode this reports a fresh session; inside `shard chat`, use `/stats` for live stats.

```bash
shard optimize
```

Compresses recent session messages into `.shard/*.md` memory files. The most useful optimizer loop runs inside `shard chat`.

## Chat Commands

Inside `shard chat`:

- `/optimize` manually compresses recent messages into categorized memory.
- `/context` shows which memory files are selected for the current focus.
- `/stats` shows session messages, optimize count, token estimates, effort, and focus.
- `/effort low|medium|high|max` changes the displayed effort level, retrieval budget, and optimizer wording.
- `/focus <category>` always includes a category such as `bugs`, `api`, `ui`, or `architecture`.
- `/exit` quits the chat session.

## How It Works

Shard runs a small compression and retrieval loop:

1. `shard init` creates local Markdown memory in `.shard/`.
2. `shard chat` loads only relevant memory files before each model call.
3. Every 10 user/assistant message pairs, Shard automatically runs optimization.
4. `/optimize` asks the model to extract durable knowledge into sections like `tasks`, `bugs`, `architecture`, `commands`, and `api`.
5. Shard appends those sections to matching `.shard/*.md` files and keeps only a short current summary in the active session.

This keeps context smaller while preserving the details that matter across a long coding session.

## Example Session

```text
$ shard init
Initialized .shard memory files.

$ shard chat
Shard chat started. Type /exit to quit.
> /focus api
Focus: api
> why is my provider failing after the first request?

Model: claude-3-5-sonnet-20241022 | Effort: medium | Context: 132 tokens

The likely issue is request construction or an unhandled non-200 response. Check that the Anthropic version header is present and log the decoded error message before returning.

> /context
Memory files:
- current.md (22 tokens)
- api.md (21 tokens)
Estimated context size: 43 tokens

> /stats
Shard stats
Session messages: 2
Optimize count: 0
Estimated tokens: 68
Current context estimate: 132
Effort: medium
Focus: api

> /optimize
Optimized memory. Before: 68 tokens | After: 31 tokens
```

## Retrieval

Shard uses simple keyword retrieval for the MVP.

- Prompts containing `bug`, `error`, `crash`, `fail`, or `fix` load `bugs.md`, `commands.md`, `codebase.md`, and `current.md`.
- Prompts containing `architecture`, `design`, or `system` load `architecture.md`, `decisions.md`, and `current.md`.
- Prompts containing `ui`, `interface`, `terminal`, or `theme` load `ui.md` and `current.md`.
- Prompts containing `api`, `provider`, `model`, `anthropic`, or `openai` load `api.md`, `dependencies.md`, and `current.md`.
- `current.md` is always loaded.
- `/focus <category>` always includes that category.

Shard does not load the whole `.shard/` folder by default.

## Token Estimation

Shard uses a deliberately simple estimate:

```text
estimatedTokens = len(text) / 4
```

It tracks current context estimate, session token estimate, and before/after optimize estimates.

## Disclaimer

Shard does not increase model context limits. It simulates larger context through compression and retrieval.

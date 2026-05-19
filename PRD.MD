# Shard PRD (Product Requirements Document)

## 1. Overview

Shard is a cross-platform CLI tool that extends AI coding workflows by adding a persistent, token-efficient memory layer using structured Markdown files.

Shard reduces token usage, improves performance, and enables long-running sessions by compressing conversations into categorized memory and retrieving only relevant context.

Shard simulates an effective ~1M token context window without increasing actual model limits.

---

## 2. Problem

Existing AI tools:
- resend full conversation context repeatedly
- waste tokens on duplicated information
- lose important context over time
- slow down as sessions grow

Result:
- high token usage
- degraded performance
- poor long-session usability

---

## 3. Solution

Shard introduces:
- automatic conversation compression
- structured markdown memory
- context-aware retrieval
- token-efficient execution

Instead of sending full history:

Conversation → Optimize → Structured Memory → Selective Retrieval → Continue

---

## 4. Goals

### Primary Goals
- Reduce token usage by 50–80%
- Maintain or improve accuracy
- Extend session lifespan significantly
- Simulate large (~1M token) context experience

### Secondary Goals
- Fast CLI experience
- Clean UI
- Cross-platform support
- High developer adoption

---

## 5. Core Concepts

### 5.1 Memory Compression
Every 10 messages:
- compress conversation into structured memory
- remove redundancy
- extract durable knowledge

### 5.2 Memory Retrieval
Before each request:
- analyze prompt
- select relevant memory
- send only required context

### 5.3 Token Efficiency
Shard trades small optimization cost for large reductions in repeated context.

---

## 6. System Architecture

Shard CLI (Go)
- Command Layer
- Session Manager
- Optimizer Engine
- Memory Store
- Retrieval Engine
- Provider Layer
- Routing Engine
- UI Renderer

---

## 7. Tech Stack

- Language: Go
- CLI Framework: Cobra
- UI: Lipgloss
- Storage: Filesystem (Markdown)
- HTTP: Go net/http
- Tokenizer: tiktoken-compatible

---

## 8. Initialization

Command:
shard init

Behavior:
- creates .shard directory
- initializes memory files

---

## 9. Memory System

### Folder Structure

.shard/
- current.md
- tasks.md
- bugs.md
- architecture.md
- decisions.md
- codebase.md
- commands.md
- dependencies.md
- api.md
- ui.md
- agents.md
- changelog.md

---

### File Format

Each file uses frontmatter:

---
category: bugs
updated: timestamp
priority: high
related: [api, commands]
---

Content example:

# Bugs

## Auth crash
- error: null pointer
- attempted fix: guard added
- status: unresolved

---

## 10. Optimization System (/optimize)

### Trigger
- automatically after every 10 messages
- manually via command

Commands:
/optimize
/optimize aggressive
/optimize max

---

### Behavior

1. collect last 10 messages
2. send to model
3. extract durable information:
   - goals
   - tasks
   - bugs
   - architecture
   - decisions
4. categorize into markdown files
5. remove duplicates
6. update current.md
7. clear redundant chat context

---

### Modes

Normal:
- balanced compression

Aggressive:
- higher compression
- minimal detail

Max:
- merges memory
- removes duplication
- restructures knowledge
- maximizes compression
- simulates ~1M token context

---

### Output Example

Optimized 10 messages

Before: 40k tokens  
After: 6k tokens  
Saved: 85%

Updated:
- tasks.md
- bugs.md
- architecture.md

Effective context: ~300k tokens

---

## 11. Optimization Prompt

You are Shard's memory optimizer.

Extract only durable, reusable knowledge:
- goals
- tasks
- bugs
- architecture decisions
- attempted fixes
- commands
- file paths
- APIs/providers
- UI decisions
- unresolved issues
- next steps

Ignore:
- repetition
- small talk
- temporary reasoning
- outdated ideas

Output updates for:
- current.md
- tasks.md
- bugs.md
- architecture.md
- decisions.md
- codebase.md
- commands.md
- dependencies.md
- api.md

---

## 12. Retrieval System

Process:
1. analyze user prompt
2. determine relevant categories
3. load relevant markdown files
4. build context

Example:

Prompt: fix auth bug

Load:
- bugs.md
- current.md
- codebase.md

---

## 13. Commands

### Core
/exit  
/optimize  
/context  
/stats  
/reset  

---

### Intelligence
/effort low|medium|high|max  
/focus <category>  
/agents <type>  

---

### Provider
/provider list  
/provider set <name>  
/provider enable <name>  
/provider disable <name>  

---

### Budget
/budget  

Only shown when requested.

---

## 14. Effort System

Levels:
- low → fast, cheap
- medium → balanced
- high → deep reasoning
- max → multi-step reasoning

Effects:
- model selection
- context size
- retrieval depth
- reasoning passes

---

## 15. Focus System

Command:
/focus <category>

Effect:
- prioritizes selected memory
- reduces context size
- improves accuracy

---

## 16. Agents

Command:
/agents <type>

Examples:
- debugger
- architect
- optimizer

Agents define:
- system prompt
- effort bias
- retrieval strategy

---

## 17. UI Design

Principles:
- minimal
- fast
- readable
- low noise

Header example:
Model: Claude Sonnet | Effort: High | Context: 18k

Thinking example:
- task: debugging
- loading bugs.md
- applying fix

---

## 18. Graph System

Command:
shard graph

Behavior:
- launches local web UI
- visualizes memory connections

Features:
- zoom and pan
- node relationships
- search and filter
- click to preview memory
- live updates

Implementation:
- Go local server
- React frontend
- Graph library (Cytoscape.js)

---

## 19. Metrics

Track:
- tokens used
- cost
- context size
- optimization count

---

## 20. Performance

Shard improves performance by:
- reducing prompt size
- eliminating repeated context
- using structured retrieval

---

## 21. Benchmark Plan

Compare:
- Shard vs baseline tools

Measure:
- tokens used
- time to complete
- accuracy

---

## 22. Token Economics

Shard introduces:
- additional optimization calls

But reduces:
- repeated context tokens

Result:
- net token savings over time

---

## 23. MVP Scope

v0.1:
- CLI chat
- markdown memory
- auto optimize
- basic retrieval

v0.2:
- effort system
- focus system
- provider support

v0.3:
- agents
- graph UI
- improved optimizer

---

## 24. Core Insight

Shard does NOT increase model context window.

It simulates large context through:
- compression
- structured memory
- selective retrieval

---

## 25. Positioning

Shard = memory layer for AI coding

---

## 26. Success Criteria

- 50–80% token reduction
- improved session length
- faster workflows
- strong developer adoption
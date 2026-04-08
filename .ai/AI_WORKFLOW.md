# AI-Assisted Development System

## Overview

This repository includes an **AI-assisted development workflow** that helps engineers plan tasks, implement features, fix errors, generate tests, and review code using AI tools such as Codex.

The system provides **structured project context** so AI can understand the codebase architecture without scanning the entire repository.

This approach improves:

* development speed
* code consistency
* architectural compliance
* token efficiency

---

# System Architecture

The AI workflow follows a layered architecture:

```
Engineer
   │
   ▼
Forge CLI
(forge plan / forge run / forge fix / forge review / forge prompt)
   │
   ▼
Prompt Layer
(.ai/prompts)
   │
   ▼
Task Context
(.ai/tasks)
   │
   ▼
Knowledge Context
(.ai/context)
   │
   ▼
AI Model
(Codex / LLM)
```

Each layer contributes specific information required for the AI to operate safely and efficiently.

---

# Repository Structure

```
.ai/
  SESSION_CONTEXT.md
  TODO.md
  README.md

  context/
    project.md
    architecture.md
    framework-go-core.md
    engineering-rules.md
    patterns.md
    rules.md
    file-index.md
    repomap.md
    feature-map.md
    symbol-map.md
    ownership.md

  prompts/
    breakdown.md
    execute.md
    fix.md
    test.md
    review.md

  tasks/
    template.md

```

---

# Context Layer

Context files provide **knowledge about the project** so AI understands the codebase structure and architecture.

Location:

```
.ai/context/
```

Files:

| File                 | Purpose                          |
| -------------------- | -------------------------------- |
| project.md           | describes the service purpose    |
| architecture.md      | defines clean architecture rules |
| framework-go-core.md | explains go-core framework usage |
| engineering-rules.md | engineering conventions          |
| patterns.md          | common code patterns             |
| rules.md             | AI modification rules            |
| file-index.md        | folder reference                 |
| repomap.md           | repository structure overview    |
| feature-map.md       | feature locations                |
| symbol-map.md        | important functions              |
| ownership.md         | code ownership boundaries        |

These files prevent AI from hallucinating architecture or modifying unrelated code.

---

# Task Layer

Tasks define the **scope of work for AI**.

Location:

```
.ai/tasks/
```

Example:

```
feature_transaction_history.md
fix_cache_api.md
```

Task files contain:

* goal
* scope layers
* allowed paths
* constraints
* expected output

Example:

```
Task: implement transaction history endpoint

Goal:
Provide API endpoint to fetch transaction history by user_id.

Scope Layers:

repository
service
handler

Allowed Paths:

internal/repository
internal/service
internal/handler
```

This ensures AI modifies only relevant parts of the repository.

---

# Prompt Layer

Prompts define **AI roles**.

Location:

```
.ai/prompts/
```

Available roles:

| Prompt       | Role             |
| ------------ | ---------------- |
| breakdown.md | planner          |
| execute.md   | backend engineer |
| fix.md       | debugger         |
| test.md      | tester           |
| review.md    | reviewer         |

These prompts control how AI processes tasks.

---

# Execution Layer

`forge` is a standalone CLI repository. Service repositories keep `.ai/` only and are discovered at runtime by walking upward from the current working directory until `.ai/` is found.

If no `.ai/` directory is found, `forge` returns:

`AI context directory (.ai) not found`

Commands:

| Command                               | Purpose                    |
| ------------------------------------- | -------------------------- |
| `forge plan <task>`                   | break down tasks           |
| `forge run <task>`                    | implement tasks            |
| `forge fix <task> --error "<msg>"`    | fix errors inside a task   |
| `forge review [<path>] [--staged]`    | review diff or path target |
| `forge prompt <task>`                 | print the final prompt     |

---

# Feature Execution Workflow

Typical development workflow:

```
create task
↓
forge plan feature_x
↓
forge run feature_x
↓
go build ./...
↓
forge fix feature_x --error "error message"
↓
forge review file.go
```

AI acts as an assistant engineer during development.

---

# Repository Awareness

The system provides AI with structural knowledge through:

* RepoMap
* FeatureMap
* SymbolMap

These allow AI to locate relevant files quickly.

Without this layer, AI would need to scan the entire repository.

---

# Prompt Assembly Contract

Forge builds task prompts in this deterministic order:

1. task context
2. feature-map.md
3. symbol-map.md
4. repomap.md
5. file-index.md
6. architecture.md
7. framework-go-core.md
8. engineering-rules.md
9. ownership.md
10. prompt template

For prompt debugging:

* `forge --print-prompt run feature_x`
* `forge --save-prompt /tmp/feature_x.prompt run feature_x`
* `forge prompt feature_x`

---

# Safety Guards

The system includes multiple safety mechanisms.

## Task Scope Guard

Each task defines:

```
Allowed Paths
```

AI may only modify files within these paths.

---

## Ownership Guard

AI must not modify code belonging to other services or frameworks.

Example:

```
go-core framework
external services
shared libraries
```

---

## Architecture Guard

AI must respect the architecture:

```
handler → service → repository
```

---

# Token Optimization Strategy

Context loading is designed to minimize token usage.

Context is injected in layers:

1. task context
2. feature map
3. symbol map
4. repository map
5. architecture context
6. framework context
7. engineering rules

This allows AI to reason about the project without reading every file.

---

# Example Task

```
.ai/tasks/feature_transaction_history.md
```

```
Task: implement transaction history endpoint

Goal:
Provide API endpoint to fetch transaction history by user_id.

Scope Layers:

repository
service
handler

Allowed Paths:

internal/repository
internal/service
internal/handler

Expected Result:

repository query
service method
handler endpoint
```

---

# Developer Workflow

Engineers follow this workflow:

```
create task
↓
ai-plan task
↓
ai-run task
↓
build project
↓
ai-fix error
↓
ai-review code
```

AI assists with planning, implementation, debugging, and review.

---

# Benefits

Advantages of this system:

* AI understands project architecture
* safer AI-assisted coding
* deterministic code modifications
* reduced token usage
* faster onboarding for new engineers

---

# Future Improvements

Planned improvements are tracked in:

```
.ai/TODO.md
```

Examples:

* AI CLI tool
* semantic code search
* automated test execution
* agent workflow loop
* CI integration
* automatic pull request generation
* multi-service awareness

---

# Conclusion

This system enables **AI-assisted backend development** while maintaining architectural safety and deterministic workflows.

It provides a structured environment where AI can help implement features, fix bugs, and review code efficiently.

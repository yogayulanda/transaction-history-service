AI Development System

This repository includes an AI-assisted development workflow designed to help engineers implement features, fix bugs, and review code using AI tools such as Codex.

The system provides structured context so AI can understand the architecture of the project without scanning the entire repository.

Folder Structure

.ai/context
Contains project knowledge and architecture information.

.ai/prompts
Defines AI roles and instructions.

.ai/tasks
Contains task definitions for AI execution.

.ai/scripts
Utility scripts that run AI workflows.

Workflow

1. Create a task

.ai/tasks/feature_example.md

2. Plan the task

ai-plan feature_example

3. Execute the task

ai-run feature_example

4. Build project

go build ./...

5. Fix errors if needed

ai-fix "error message"

6. Review code

ai-review path/to/file.go

Task Naming Convention

feature_*
fix_*
refactor_*
test_*

This determines which AI prompt will be used.

Context Layers

Project Context
Architecture Context
Framework Context
Repository Map
Feature Map
Symbol Map
Engineering Rules
Ownership Rules

These layers allow AI to understand the repository structure quickly.

Safety Guards

The system prevents AI from modifying files outside the defined task scope.

Tasks must define:

Allowed Paths

Only files within those paths may be modified.

Future Improvements

See:

.ai/TODO.md
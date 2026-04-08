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

Workflow

1. Create a task

.ai/tasks/feature_example.md

2. Plan the task

forge plan feature_example

3. Execute the task

forge run feature_example

4. Build project

go build ./...

5. Fix errors if needed

forge fix feature_example --error "error message"

6. Review code

forge review path/to/file.go

7. Inspect the final prompt if needed

forge prompt feature_example

Task Naming Convention

feature_*
fix_*
refactor_*
test_*

This determines which AI prompt will be used when explicit task metadata is not present.

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

Repository Detection

The standalone `forge` CLI starts from the current working directory and walks upward until it finds `.ai/`.

If `.ai/` is not found, it returns:

AI context directory (.ai) not found

Prompt Debugging

Use:

forge --print-prompt run feature_example
forge --save-prompt /tmp/feature_example.prompt run feature_example
forge prompt feature_example

Safety Guards

The system prevents AI from modifying files outside the defined task scope.

Tasks must define:

Allowed Paths

Only files within those paths may be modified.

Future Improvements

See:

.ai/TODO.md

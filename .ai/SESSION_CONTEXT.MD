AI Engineering System Context

Project type:
Go microservices.

Framework:
go-core.

Service:
transaction-history-service.

Workspace structure:

projects/
  go-core/
  transaction-history-service/

Architecture:

handler → service → repository

handler:
transport layer

service:
business logic

repository:
database access

Framework responsibilities (go-core):

- bootstrap
- config
- database
- logger
- lifecycle
- grpc
- http gateway
- observability

Infrastructure must come from go-core.

Purpose of AI system:

Assist engineers with:

- planning tasks
- implementing features
- fixing build errors
- generating tests
- reviewing code

Local AI configuration:

.ai/context
.ai/prompts

Context files provide project knowledge.

Prompt files define AI roles.

Token strategy:

LEVEL 1
prompt only

LEVEL 2
prompt + core context

LEVEL 3
full context

Engineer workflow:

build
↓
fix errors with AI
↓
plan feature
↓
execute feature
↓
review code
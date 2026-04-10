Code Ownership Map

Purpose:
Identify which module or service owns specific parts of the codebase.

This prevents AI from modifying code that belongs to other services or frameworks.

Ownership Rules

transaction-history-service owns:

internal/handler
internal/service
internal/repository
internal/domain

transaction-history-service must NOT modify:

go-core framework
external services
shared libraries

Framework Ownership

go-core owns:

bootstrap logic
config loading
database initialization
logging
grpc server
http gateway

Services must not modify framework internals.

Future Microservices

user-service

ownership:
internal/user/*

payment-service

ownership:
internal/payment/*

transaction-history-service

ownership:
internal/handler/*
internal/service/*
internal/repository/*
internal/domain/*

Rules

only modify code owned by the current service.

never modify code belonging to another service.
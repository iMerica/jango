# AGENTS.md

Guidance for coding agents and contributors working in this repository.

## Architecture priorities

JanGO is a batteries-included, Django-inspired web framework for Go. Preserve the product shape: apps, settings, URLs, middleware, models, migrations, QuerySets, auth, sessions, security, templates, forms, static files, admin, and REST should feel like one coherent framework.

Do not turn this into a tiny router. The goal is not a small HTTP helper or a bring-your-own-everything starter kit. The goal is a cohesive framework where the major systems understand each other.

Prefer framework-owned public APIs over leaking third-party concepts. It is fine to use strong Go libraries internally, but users should see JanGO concepts: models, querysets, migrations, settings, apps, serializers, views, permissions, and commands.

## Design principles

- Batteries included does not mean chaos included.
- Preserve Django's workflow where it matters.
- Use Go's strengths where Python's tricks do not translate.
- Make the happy path obvious.
- Make escape hatches boring.
- Prefer generated metadata over clever reflection.
- Keep public APIs stable enough for migrations, serializers, admin, and tests to build on.

## Codegen over runtime magic

Go cannot dynamically import arbitrary app packages the way Django can import Python modules. Prefer explicit registration plus source generation over runtime import tricks, plugin-heavy discovery, or reflection-only metadata.

Generated registries should own installed app wiring, model registration, command discovery, route discovery, serializers, admin registration, and checks where possible. Runtime behavior should be inspectable and deterministic.

## PostgreSQL-first as an implementation constraint

PostgreSQL is the first-class database target for the MVP. Do not weaken the ORM or migration design to chase lowest-common-denominator SQL too early.

PostgreSQL-specific capabilities are allowed and expected: JSONB, arrays, UUIDs, full-text search, partial indexes, constraints, extensions, transactional DDL where available, and rich schema editor behavior.

Other database backends can come later behind the same abstractions.

## REST is built in, not optional

The REST stack is part of the framework, not a plugin. Design ORM, auth, permissions, serializers, pagination, filtering, content negotiation, metadata, and schema generation so API work feels native.

DRF is the conceptual reference point. Do not scatter API behavior across unrelated packages in ways that make viewsets, routers, serializers, permissions, and schema generation hard to reason about.

## Security defaults are required

Security belongs in the core framework contract. Keep CSRF, signed cookies, password hashing, sessions, host validation, secure headers, clickjacking protection, deploy checks, safe template defaults, and SQL parameterization as first-class concerns.

Avoid APIs that encourage string-built SQL, mutable global settings after setup, insecure production defaults, or framework-owned flows that skip security middleware.

## Testing expectations

Tests should cover framework behavior, not just isolated helpers. Prefer end-to-end-ish coverage for app startup, URL resolution, middleware ordering, model metadata, migration state, queryset SQL, auth/session behavior, CSRF, serializers, and generic views.

Use focused unit tests where the behavior is local, but add integration tests when a change crosses subsystem boundaries. For parsers, routers, CSRF, auth headers, serializer validation, and SQL expression compilation, fuzz tests are welcome once the surface stabilizes.

## Build, test, and lint commands

Use these commands from the repository root:

```bash
go test ./...
go vet ./...
go build ./cmd/jango
gofmt -w <changed-go-files>
```

If the environment cannot write to the default Go build cache, use a repo-local cache:

```bash
mkdir -p .gocache
GOCACHE="$(pwd)/.gocache" go test ./...
GOCACHE="$(pwd)/.gocache" go vet ./...
GOCACHE="$(pwd)/.gocache" go build ./cmd/jango
```

For narrow feedback during ORM or migration work:

```bash
go test ./orm ./migrations
```

## Planned package shape

```text
cmd/
  jango/

apps/
conf/
http/
urls/
views/
middleware/
models/
orm/
db/
migrations/
auth/
sessions/
security/
rest/
templates/
forms/
staticfiles/
checks/
signals/
cache/
admin/
contrib/
examples/
```

This layout is intentionally framework-shaped. It follows Go's package model, but the mental model is Django.

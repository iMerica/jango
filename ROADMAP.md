# Roadmap

JanGO is early. The API will change while the core systems settle.

```text
JanGO core progress

[##########----------] 50%

Done-ish:
- Project direction
- App concept
- Settings shape
- Model declaration style
- Migration command shape
- Middleware design
- REST framework design

In progress:
- ORM internals
- Migration graph
- PostgreSQL schema editor
- QuerySet semantics
- Auth and sessions
- Security middleware
- Serializers and ViewSets

Not done:
- Admin
- Full template integration
- Full forms layer
- Production hardening
- Compatibility test matrix
```

## Subsystem Status

| Area | Status | Notes |
| --- | --- | --- |
| Project scaffolding | Planned | `startproject`, `startapp`, generated layout |
| Apps | Planned | Django-style installed apps, Go-native registration |
| Settings | Planned | Typed config with environment overlays |
| Views | Planned | Function views, class-style views, generic views |
| URL routing | Planned | URLconf-style route maps, namespacing, reverse lookup |
| Middleware | Planned | Onion model, request hooks, response hooks, exception hooks |
| Models | In design | Declarative model metadata with Go-native types |
| QuerySets | In design | Lazy, immutable query builders |
| Migrations | In design | `makemigrations`, `migrate`, graph, schema editor |
| PostgreSQL | Planned | First-class backend, not lowest-common-denominator SQL |
| REST framework | In design | Serializers, API views, viewsets, routers, permissions |
| Auth | Planned | Users, groups, permissions, auth backends |
| Sessions | Planned | Cookie, DB, and cache-backed sessions |
| Security | Planned | CSRF, secure headers, host validation, signed cookies |
| Admin | Later | Planned after ORM, forms, auth, and templates settle |

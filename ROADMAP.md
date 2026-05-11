# Roadmap

JanGO is early. The API will change while the core systems settle.

```text
JanGO core progress

[##############------] 70%

Done-ish:
- Project direction
- App concept
- Settings shape
- Model declaration style
- Migration command shape
- Middleware design
- REST framework design
- Live PostgreSQL ORM integration
- Auth/session/security pass

In progress:
- Serializers and ViewSets
- REST filtering, pagination, metadata, and schema generation
- Forms/admin/test-client integration

Not done:
- Full template integration
- Full forms layer
- Admin
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
| Models | Done-ish | Struct/tag metadata through `orm.ModelMeta`, normalized DB columns |
| QuerySets | Done-ish | Lazy typed querysets backed by parameterized PostgreSQL SQL |
| Migrations | Done-ish | `makemigrations`, `migrate`, graph, schema editor; needs many-to-many and postgres extension polish |
| PostgreSQL | Done-ish | First-class backend through framework ORM/DB layer |
| REST framework | In progress | Placeholder package; next incomplete plan is serializers, API views, viewsets, routers, permissions |
| Auth | Done-ish | Users, groups, permissions, auth backends; verify against current ORM API |
| Sessions | Done-ish | Cookie, DB, and cache-backed sessions |
| Security | Done-ish | CSRF, secure headers, host validation, signed cookies |
| Admin | Later | Planned after ORM, forms, auth, and templates settle |

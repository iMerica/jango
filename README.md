# JanGO

> The web framework for perfectionists with deadlines who want to use Go.

[![Status](https://img.shields.io/badge/status-pre_alpha-orange)](#project-status)
[![API](https://img.shields.io/badge/api-unstable-red)](#project-status)
[![Go](https://img.shields.io/badge/language-Go-00ADD8)](#jango)
[![Batteries Included](https://img.shields.io/badge/style-batteries_included-black)](#omakase-not-a-la-carte)
[![License](https://img.shields.io/badge/license-MIT-lightgrey)](#license)

JanGO is Django's worldview rebuilt for Go. Not a router. Not an ORM with vibes. It is a batteries-included web framework with models, migrations, QuerySets, views, URL routing, middleware, settings, apps, auth, sessions, security, PostgreSQL, and a built-in REST framework.

Go compiles to a single binary, starts fast, deploys easily, and has a strong standard library. What it does not have much of is the luxurious, cohesive, "I have a product to ship" lane. JanGO is for people who like Go, but also like nice things.

Sometimes you do not want to pick a router, an ORM, a migration tool, auth, sessions, CSRF, config, pagination, serializers, and OpenAPI tooling before you can ship the actual product. Django is omakase: the pieces are chosen to work together. JanGO is trying to bring that idea to Go.

Also, vibe coding makes this more useful, not less. A high-level backend gives both humans and AI a shared shape: models, migrations, views, serializers, permissions, sessions. That means fewer tokens spent re-explaining plumbing, fewer bespoke stacks for the model to hallucinate, and more time spent on the product.

PostgreSQL is the first database target, and security defaults are part of the framework contract.

## Django vs JanGO

Same idea, Go shape.

<table>
<tr>
<th>Django</th>
<th>JanGO</th>
</tr>
<tr>
<td>

```python
class Topping(models.Model):
    name = models.CharField(max_length=80, unique=True)
    is_vegetarian = models.BooleanField(default=True)

    class Meta:
        ordering = ["name"]


class Pizza(models.Model):
    name = models.CharField(max_length=120, unique=True)
    toppings = models.ManyToManyField(
        Topping,
        related_name="pizzas",
        blank=True,
    )

    class Meta:
        ordering = ["name"]
```

</td>
<td>

```go
type Topping struct {
	ID           uint   `jango:"primary_key"`
	Name         string `jango:"type:char,max_length:80,unique"`
	IsVegetarian bool   `jango:"type:boolean,default:true"`
}

func (Topping) Meta() orm.ModelOptions {
	return orm.ModelOptions{Ordering: []string{"name"}}
}

type Pizza struct {
	ID       uint       `jango:"primary_key"`
	Name     string     `jango:"type:char,max_length:120,unique"`
	Toppings []*Topping `jango:"related_name:pizzas"`
}

func (Pizza) Meta() orm.ModelOptions {
	return orm.ModelOptions{Ordering: []string{"name"}}
}

func init() {
	orm.RegisterModel("pizza", &Topping{})
	orm.RegisterModel("pizza", &Pizza{})
}
```

</td>
</tr>
</table>

That model declaration should feed migrations:

```bash
jango makemigrations
jango migrate
```

And then give you a QuerySet-style API:

```go
pizzas, err := orm.Objects[Pizza]("pizza", "Pizza").
    Filter(orm.L("toppings__name", "Mozzarella")).
    OrderBy("name").
    AllRecords(ctx)
if err != nil {
    return err
}
```

The goal is not to make Go look like Python. The goal is to make Go feel like a complete web framework instead of a trip to Home Depot.

## Project status

This is early. Do not put your bank on it yet. The API will change.

```text
JanGO core progress
[##########----------] 50%
```

See [ROADMAP.md](ROADMAP.md) for the current subsystem status.

## Omakase, not a la carte

Small packages and clear boundaries are real strengths. The point is not to fight Go. The point is to stop making every product team assemble the same stack from scratch.

| A la carte Go stack | JanGO |
| --- | --- |
| Pick a router | URLconf-style routing included |
| Pick an ORM | Models and QuerySets included |
| Pick migrations | `makemigrations` and `migrate` included |
| Pick validation | Forms and serializers included |
| Pick auth | Auth, permissions, and sessions included |
| Pick API tooling | REST framework included |
| Pick security middleware | Security defaults included |
| Pick project layout | Project and app conventions included |
| Pick admin tooling | Admin planned as a contrib app |

Minimalism is great until you are building the same login flow for the ninth time.

## Isn't this Anti-Go?

Yes. A little.

JanGO is not trying to be the smallest possible Go library or a shrine to explicit plumbing. It uses Go for what Go is great at: simple deployment, fast binaries, good concurrency, and readable code. But it does not treat "assemble every screw yourself" as a product requirement. If AI pushes more code toward generated implementation detail, then some sacred Go idioms are suddenly fair game again. The trade is a bit less ceremony-free minimalism for a lot more done.

## Django in spirit, not internals

JanGO is not a one-to-one source port. That would be weird. Also bad.

Django is Python. Python gives you runtime imports, metaclasses, descriptors, introspection, and all kinds of spooky little tricks that are very useful when you are building a framework. Go is not that. Go is compiled. Go likes explicitness. Go does not want you doing magic in a closet.

So JanGO ports the ideas, not the mechanism.

| Django concept | Python makes this easy with | JanGO does this with | Why |
| --- | --- | --- | --- |
| `INSTALLED_APPS` | Runtime imports | Explicit registration plus codegen | Go cannot import packages dynamically at runtime |
| Model classes | Metaclasses and descriptors | Model declarations plus generated metadata | Rich metadata without insane struct tags |
| QuerySets | Dynamic method chains | Immutable typed query builders | Same lazy query feel, Go-native mechanics |
| `manage.py` | Script entrypoint | `jango` CLI | Same workflow, compiled binary |
| Migrations | Runtime model inspection | State snapshots and generated migration files | Deterministic diffs |
| Middleware | Callable onion | Interface-based onion | Same lifecycle, explicit hooks |
| DRF serializers | Introspection | Metadata-backed serializers | Same role, static-friendly implementation |
| Admin autodiscovery | Import side effects | Generated registries | Same UX, compiled implementation |
| Settings module | Module globals | Typed settings with env overlays | Safer under concurrency |
| Reusable apps | Python packages | Go packages with app configs | Same product model, different internals |

If you know Django, you should know what JanGO is trying to do immediately. If you know Go, you should be able to read the implementation without yelling.

That is the balance.

## Models, migrations, and the luxury layer

The best part of Django is that its systems talk to each other. Models are the source of truth for schema, migrations, the query API, and enough metadata to generate full CRUD views. Apps wire into settings. Auth depends on models. Admin depends on models, forms, templates, permissions, and URLs. Django REST Framework sits on top of views, auth, serializers, pagination, filtering, and schema generation.

That is the luxury layer. JanGO is built around the same idea.

A model should be more than a data container. It should be the source of truth for:

- Database schema
- Migrations
- QuerySets
- Forms
- Serializers
- Admin screens
- Validation
- Permissions
- OpenAPI schema

## Built-in REST framework

REST is not a plugin here. Django REST Framework works because it is a framework on top of a framework. It has serializers, API views, generic views, viewsets, routers, permissions, authentication, throttling, filtering, pagination, content negotiation, metadata, and schema generation.

JanGO should have that from the beginning.

Target shape:

```go
var PizzaSerializer = rest.ModelSerializer(Pizza,
    rest.Fields("id", "name", "toppings"),
)

type PizzaViewSet struct {
    rest.ModelViewSet
}

func NewPizzaViewSet() *PizzaViewSet {
    return &PizzaViewSet{
        ModelViewSet: rest.ModelViewSet{
            QuerySet:    Pizza.Objects().All(),
            Serializer:  PizzaSerializer,
            Permissions: []rest.Permission{rest.IsAuthenticatedOrReadOnly()},
        },
    }
}
```

Router target:

```go
router := rest.NewRouter()
router.Register("pizzas", NewPizzaViewSet())

urlpatterns := urls.Patterns(
    urls.Include("/api/", router.URLs()),
)
```

Again, not a literal Python port. The shape is what matters.

## Principles

- Batteries included, but still Go-shaped.
- Django workflow where it matters; Go implementation where it helps.
- Familiar places for Django developers; explicit, inspectable behavior for Go developers.
- Models, migrations, QuerySets, forms, admin, and APIs should talk to each other.
- Codegen over runtime magic.
- PostgreSQL first, not lowest-common-denominator SQL first.
- Security and REST are core framework concerns.

## Planned command surface

```bash
jango startproject mysite
jango startapp blog

jango runserver
jango check
jango check --deploy
jango makemigrations
jango migrate
jango showmigrations
jango sqlmigrate blog 0001
jango createsuperuser
jango collectstatic
jango test
```

## Compatibility promise

JanGO does aim to preserve the parts that matter:

| Django behavior | JanGO goal |
| --- | --- |
| Installed apps | Yes |
| App registry | Yes |
| Settings | Yes |
| URLconf | Yes |
| Views | Yes |
| Middleware onion | Yes |
| Models | Yes |
| QuerySets | Yes |
| Migrations | Yes |
| Management commands | Yes |
| Auth and permissions | Yes |
| Sessions | Yes |
| CSRF | Yes |
| Security middleware | Yes |
| Static files | Yes |
| Templates | Yes |
| Forms | Yes |
| Admin | Planned |
| REST framework | Yes, built in |
| Third-party Django app compatibility | No |
| Django workflow familiarity | Yes |

## License

MIT.

## Contributing

This project is early. The best contributions right now tighten the core shape: models, migrations, QuerySets, app registry behavior, REST primitives, and compatibility tests against Django-like workflows.

See [AGENTS.md](AGENTS.md) for architecture notes and working conventions.

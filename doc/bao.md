# bao

Package `bao` provides generic CRUD helpers built on top of
[bun](https://bun.uptrace.dev/). It handles the repetitive parts of
database access—transactions, existence checks, related-model persistence,
UUID validation—while leaving query customisation to the caller via a
`queryFn` callback.

## Functions

### SelectQuery

```go
func SelectQuery[ModelT any](ctx context.Context, db bun.IDB, model *ModelT) (*bun.SelectQuery, *schema.Table, error)
```

Creates a `bun.SelectQuery` for `model` and returns the resolved
`schema.Table`. `ModelT` must be a struct or a slice; returns
`ErrModelNotStructOrSlice` otherwise.

### SelectForUpdateQuery

```go
func SelectForUpdateQuery[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, skipLocked bool) (*bun.SelectQuery, *schema.Table, error)
```

Like `SelectQuery` but appends `FOR UPDATE OF <alias> [SKIP LOCKED]`.
Useful for pessimistic locking in queue-style workloads.

### Find

```go
func Find[ModelT any](ctx context.Context, db bun.IDB, queryFn func(q *bun.SelectQuery)) ([]*ModelT, error)
```

Returns all rows matching the optional `queryFn` filter. Pass `nil` to
return every row.

### FindFirst

```go
func FindFirst[ModelT any](ctx context.Context, db bun.IDB, queryFn func(q *bun.SelectQuery)) (*ModelT, error)
```

Like `Find` but scans into a single struct. Returns an error if no row is
found (bun's `sql.ErrNoRows` propagates).

### FindByID

```go
func FindByID[ModelT any](ctx context.Context, db bun.IDB, id string, queryFn func(q *bun.SelectQuery)) (*ModelT, error)
```

Fetches the row whose single primary key equals `id`. `id` must be a valid
UUID string; returns `ErrIDNotUUID` otherwise. Returns `ErrOnePrimaryKey`
when the table has zero or more than one PK.

### FindByIDForUpdate

```go
func FindByIDForUpdate[ModelT any](ctx context.Context, db bun.IDB, id string, skipLocked bool, queryFn func(q *bun.SelectQuery)) (*ModelT, error)
```

Combines `FindByID` with `FOR UPDATE OF <alias> [SKIP LOCKED]`.

### Create

```go
func Create[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, befores []hook.Before[ModelT], afters []hook.After[ModelT]) error
```

Inserts `model` inside a transaction. Runs `befores` hooks before the
INSERT and `afters` hooks (outside the transaction) on success. Also
persists any relations tagged `bao:"persist"`.

### Update

```go
func Update[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, befores []hook.Before[ModelT], afters []hook.After[ModelT]) error
```

Updates `model` by primary key inside a transaction. Returns
`ErrUpdateNotExists` if the row does not exist before the UPDATE runs.
Same hook and relation semantics as `Create`.

### Delete

```go
func Delete[ModelT any](ctx context.Context, db bun.IDB, model *ModelT, queryFn func(q *bun.DeleteQuery), befores []hook.Before[ModelT], afters []hook.After[ModelT]) error
```

Deletes `model` inside a transaction. If `queryFn` is `nil` the delete
uses `WherePK()`. Also deletes any relations tagged `bao:"persist"`.

### Trx

```go
func Trx(ctx context.Context, db bun.IDB, fn func(ctx context.Context, tx bun.IDB) error) error
```

Wraps `fn` in a transaction. If `db` is already a `bun.Tx` the existing
transaction is reused (no savepoint). Rolls back automatically on error,
commits on success.

## Errors

| Sentinel | Meaning |
|----------|---------|
| `ErrModelNotStructOrSlice` | `ModelT` must be a struct or a slice |
| `ErrModelNotStruct` | `ModelT` must be a struct (for write operations) |
| `ErrOnePrimaryKey` | Table must have exactly one PK for ID-based lookups |
| `ErrUpdateNotExists` | Row does not exist when `Update` is called |
| `ErrIDNotUUID` | Supplied ID is not a valid UUID string |

## Relation persistence (`bao:"persist"`)

When `Create`, `Update`, or `Delete` runs, `bao` iterates the bun-defined
relations of the model. For any relation field tagged `bao:"persist"`, the
join table rows tied to the parent PK are deleted first, then re-inserted
from the current in-memory value (for create/update). This implements a
simple replace-all strategy for has-many / m2m associations.

## Hooks

Package `bao/hook` defines the hook function signatures used by the write
operations.

```go
// Before is called inside the transaction before the DB write.
// Return a non-nil error to abort.
type Before[ModelT any] func(ctx context.Context, db bun.IDB, model *ModelT) error

// After is called outside the transaction after a successful write.
type After[ModelT any] func(ctx context.Context, model *ModelT)
```

### LocalParameterBeforeHook

```go
func LocalParameterBeforeHook[ModelT any](fn func(ctx context.Context) map[string]string) Before[ModelT]
```

Returns a `Before` hook that runs `SET LOCAL "key" = value` for each
entry in the map returned by `fn`. Useful for setting PostgreSQL
session-local configuration (e.g. `app.current_user_id`) inside RLS
policies before a write.

## Example

```go
type User struct {
    bun.BaseModel `bun:"users"`
    ID    string `bun:"id,pk"`
    Email string `bun:"email"`
}

// Fetch a single user by UUID.
user, err := bao.FindByID[User](ctx, db, userID, func(q *bun.SelectQuery) {
    q.Where("deleted_at IS NULL")
})

// Create with a before hook that stamps audit metadata.
err = bao.Create(ctx, db, &User{ID: uuid.NewString(), Email: "x@example.com"},
    []hook.Before[User]{auditBeforeHook},
    nil,
)
```

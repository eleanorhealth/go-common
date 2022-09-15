# bao

bao provides a few additional features for [Bun](https://github.com/uptrace/bun):

* Helper functions with type parameters for finding, saving, and deleting models
* Transaction-aware before and after hooks for create, update, and delete operations
* "Nested transaction" handling

## Basic Usage

```go
type user struct {
    ID        string `bun:",pk"`
    NameFirst string
    NameLast  string
}

newUser := &user{
    ID:        uuid.New().String(),
    NameFirst: "John",
    NameLast:  "Smith",
}

// db is a *bun.DB
err := bao.Create(context.Background(), db, newUser, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Hooks

```go
type user struct {
    ID        string `bun:",pk"`
    NameFirst string
    NameLast  string
}

myBeforeSaveHook := func(ctx context.Context, db bun.IDB, model *user) error {
    fmt.Println("before save")
    return nil
}

myAfterSaveHook := func(ctx context.Context, model *user) {
    fmt.Println("after save")
}

newUser := &user{
    ID:        uuid.New().String(),
    NameFirst: "John",
    NameLast:  "Smith",
}

// db is a *bun.DB
err := bao.Create(
    context.Background(),
    db,
    newUser,
    []hook.Before[user]{myBeforeSave},
    []hook.After[user]{myAfterSave},
)
if err != nil {
    log.Fatal(err)
}
```

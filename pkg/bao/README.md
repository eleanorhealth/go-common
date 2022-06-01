# bao

## Basic Usage

```go
type user struct {
    ID        string `bun:",pk"`
    NameFirst string
    NameLast  string
}

// db is a *bun.DB
userStore := bao.NewStore[user](db)

newUser := &user{
    ID:        uuid.New().String(),
    NameFirst: "John",
    NameLast:  "Smith",
}

err := userStore.Save(context.Background(), newUser)
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

// db is a *bun.DB
userStore := bao.NewStore[user](db)
userStore.WithBeforeSaveHooks(myBeforeSaveHook)
userStore.WithAfterSaveHooks(myAfterSaveHook)

newUser := &user{
    ID:        uuid.New().String(),
    NameFirst: "John",
    NameLast:  "Smith",
}

err := userStore.Save(context.Background(), newUser)
if err != nil {
    log.Fatal(err)
}
```

# env

## Basic Usage

```go
os.Setenv("FOO", "bar")

fmt.Println(env.Get("FOO", ""))                  // bar
fmt.Println(env.Get("DOES_NOT_EXIST", "foobar")) // foobar
fmt.Println(env.Get("FOO", []byte{}))            // [98 97 114]

env.Setenv("local")
fmt.Println(env.IsLocal()) // true

env.Setenv("qa")
fmt.Println(env.IsQA()) // true

env.Setenv(env.EnvProd)
fmt.Println(env.IsProd()) // true

env.Setenv("unknown") // panic("invalid env \"unknown\"")
```

> Note: an invalid env provided to `env.Setenv()` results in a panic. Valid envs are `local`, `qa`, and `prod`.

> Note: Calling `env.IsLocal|IsQA|IsProd()` prior to `env.Setenv()` with a valid env results in a panic.

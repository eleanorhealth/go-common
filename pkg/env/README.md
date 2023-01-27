# env

## Basic Usage

```go
os.Setenv("FOO", "bar")

fmt.Println(env.Get("FOO", ""))                  // bar
fmt.Println(env.Get("DOES_NOT_EXIST", "foobar")) // foobar
fmt.Println(env.Get("FOO", []byte{}))            // [98 97 114]

env.Setenv("local")
fmt.Println(env.IsLocal()) // true

env.Setenv(env.Env("qa"))
fmt.Println(env.IsQA()) // true

env.Setenv(env.EnvProd)
fmt.Println(env.IsProd()) // true

env.Setenv("unknown") // panic("invalid env \"unkown\"")
```

> Note: an invalid env provided to `env.SetEnv()` results in a panic. Valid envs are `local`, `qa`, and `prod`.

> Note: Calling `env.IsLocal|IsQA|IsProd()` prior to `env.SetEnv()` with a valid env results in a panic.

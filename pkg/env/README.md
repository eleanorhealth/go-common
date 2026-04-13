# env

## Basic Usage

```go
os.Setenv("FOO", "bar")

fmt.Println(env.Get("FOO", ""))                  // bar
fmt.Println(env.Get("DOES_NOT_EXIST", "foobar")) // foobar
fmt.Println(env.Get("FOO", []byte{}))            // [98 97 114]

token := env.Setenv("local")
fmt.Println(env.IsLocal(token)) // true

token = env.Setenv("qa")
fmt.Println(env.IsQA(token)) // true

token = env.Setenv(env.EnvProd)
fmt.Println(env.IsProd(token)) // true

env.Setenv("unknown") // panic: invalid env "unknown"
```

> Note: an invalid env provided to `env.Setenv()` results in a panic. Valid envs are `local`, `qa`, and `prod`.

> Note: `env.IsLocal`, `env.IsQA`, and `env.IsProd` require the value returned by `env.Setenv` — calling them without it is a compile-time error.

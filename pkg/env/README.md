# env

## Basic Usage

```go
os.Setenv("FOO", "bar")

fmt.Println(env.Get("FOO", ""))                  // bar
fmt.Println(env.Get("DOES_NOT_EXIST", "foobar")) // foobar
fmt.Println(env.Get("FOO", []byte{}))            // [98 97 114]

os.Setenv("ENV", "local")
fmt.Println(env.IsLocal()) // true

os.Setenv("ENV", "qa")
fmt.Println(env.IsQA()) // true

os.Setenv("ENV", "prod")
fmt.Println(env.IsProd()) // true
```

# errs

## Basic Usage

```go
var myError = errors.New("my error")

err := errs.Wrap(myError, "this is a wrapped error")
fmt.Println(errs.Cause(err) == myError) // true

fmt.Println(errors.Is(err, myError)) // true
```

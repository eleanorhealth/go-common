# request

go request library

## Usage

```go
client := request.NewClient("https://jsonplaceholder.typicode.com/todos")

todos := []struct {
    ID        int    `json:"id"`
    UserID    int    `json:"userId"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}{}

values := url.Values{}
values.Add("userId", "2")

res, err := client.Get(context.Background(), "/", values, &todos)
if err != nil {
    log.Fatal(err)
}

fmt.Println(res)
fmt.Println(todos)
```

### Custom http client

```go
httpClient := &http.Client{
    Timeout: time.Second * 5,
}

client := request.NewClient(
    "https://jsonplaceholder.typicode.com/todos",
    request.WithHTTPClient(httpClient),
)
```

### Custom error handling

```go
type CustomError struct {
	Detail string `json:"detail"`
}

func (c *CustomError) Error() string {
	return fmt.Sprintf("custom error details: %s", c.Detail)
}

client := request.NewClient(
    "https://jsonplaceholder.typicode.com/todos",
    request.WithErrChecker(func(req *http.Request, res *http.Response) error {
        if res.StatusCode != http.StatusOK { // your custom error handling here...
            b, err := io.ReadAll(res.Body)
            if err != nil {
                return err
            }

            custErr := &CustomError{}
            err = json.Unmarshal(b, custErr)
            if err != nil {
                return err
            }

            return custErr
        }

        return nil
    }),
)

items := []struct {
    // ...
}{}
_, err := client.Get(context.Background(), "/", nil, &items)
if err != nil {
    if custErr, ok := err.(*CustomError); ok {
        log.Fatal(custErr.Detail)
    }
}

fmt.Println(items)
```

### Token auth

```go
client := request.NewClient(
    "https://some-token-authed-api.com",
    request.WithTokenAuth("Bearer <token-here>"),
)
client = request.NewClient(
    "https://some-token-authed-api.com",
    request.WithTokenAuth("Token <token-here>"),
)
client = request.NewClient(
    "https://some-token-authed-api.com",
    request.WithTokenAuth("<token-here>"),
)
```

### Basic auth

```go
client := request.NewClient(
    "https://some-basic-authed-api.com",
    request.WithBasicAuth("user", "pass"),
)
```

### All together

```go
customHTTPClient := &http.Client{
    Timeout: time.Second * 5,
}

client := request.NewClient(
    "https://some-token-authed-api.com",
    request.WithHTTPClient(customHTTPClient),
    request.WithTokenAuth("Bearer <token-here>"),
    request.WithErrChecker(func(req *http.Request, res *http.Response) error {
        if res.StatusCode != http.StatusOK {
            return fmt.Errorf("some error occurred %d %s%s", res.StatusCode, req.URL.Host, req.URL.Path)
        }

        return nil
    }),
)
```